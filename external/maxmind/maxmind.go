// Package maxmind keeps a local copy of the MaxMind GeoLite2 ASN database
// in PostgreSQL, mapping IP networks to autonomous system numbers and
// organisations. It is used to annotate request logs with network data
// without storing any per-request geolocation lookups.
package maxmind

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"path"
	"strconv"
	"time"
)

var (
	maxmindAccountID  = flag.String("maxmind-account-id", "", "MaxMind account ID used to download GeoLite2 databases")
	maxmindLicenseKey = flag.String("maxmind-license-key", "", "MaxMind license key used to download GeoLite2 databases")
)

const (
	downloadURL = "https://download.maxmind.com/geoip/databases/GeoLite2-ASN-CSV/download?suffix=zip"

	refreshInterval = 48 * time.Hour
	refreshTimeout  = 10 * time.Minute

	// The GeoLite2-ASN CSV archive is around 10MB; the limits below are
	// generous bounds to stop a bad response exhausting memory.
	maxDownloadBytes = 64 << 20
)

// RegisterGoroutine returns a worker that downloads the GeoLite2 ASN
// database at startup and then every two days, replacing the contents of
// the local asn_networks table. If credentials are not configured it does
// nothing; the site keeps working, just without ASN data.
func RegisterGoroutine(ctx context.Context) func() {
	if *maxmindAccountID == "" || *maxmindLicenseKey == "" {
		return func() {
			slog.Info("MaxMind credentials not configured, ASN data will not be updated")
		}
	}
	return func() {
		refreshASNData(ctx)

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				refreshASNData(ctx)
			}
		}
	}
}

func refreshASNData(ctx context.Context) {
	last, err := lastRefresh(ctx)
	if err != nil {
		slog.Error("Failed to check ASN refresh timestamp", "error", err)
		return
	}
	if !last.IsZero() && time.Since(last) < refreshInterval {
		return
	}
	slog.Info("Refreshing GeoLite2 ASN data")
	ctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()

	networks, err := downloadASNCSV(ctx)
	if err != nil {
		slog.Error("Failed to download GeoLite2 ASN data", "error", err)
		return
	}

	if err := replaceASNNetworks(ctx, networks); err != nil {
		slog.Error("Failed to update ASN networks", "error", err)
		return
	}
	slog.Info("GeoLite2 ASN data updated", "networks", len(networks))
}

func downloadASNCSV(ctx context.Context) ([]asnNetwork, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating download request: %w", err)
	}
	// Credentials are sent in the Authorization header rather than the URL
	// so they cannot leak into logs or error messages.
	req.SetBasicAuth(*maxmindAccountID, *maxmindLicenseKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("unexpected status %d downloading GeoLite2 ASN database: %s", resp.StatusCode, body)
	}

	archive, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading download response: %w", err)
	}
	if len(archive) > maxDownloadBytes {
		return nil, fmt.Errorf("download exceeds maximum size of %d bytes", maxDownloadBytes)
	}

	return parseASNZip(archive)
}

type asnNetwork struct {
	network      netip.Prefix
	asn          int64
	organization string
}

func parseASNZip(archive []byte) ([]asnNetwork, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("opening GeoLite2 ASN archive: %w", err)
	}

	var networks []asnNetwork
	seen := map[string]bool{}
	for _, file := range reader.File {
		base := path.Base(file.Name)
		if base != "GeoLite2-ASN-Blocks-IPv4.csv" && base != "GeoLite2-ASN-Blocks-IPv6.csv" {
			continue
		}

		if file.UncompressedSize64 > maxDownloadBytes {
			return nil, errors.New("ASN CSV exceeds maximum size")
		}
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("opening %s: %w", base, err)
		}
		data, readErr := io.ReadAll(io.LimitReader(rc, maxDownloadBytes+1))
		rc.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading %s: %w", base, readErr)
		}
		if uint64(len(data)) != file.UncompressedSize64 || len(data) > maxDownloadBytes {
			return nil, errors.New("truncated or oversized ASN CSV")
		}
		parsed, err := parseASNCSV(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", base, err)
		}
		networks = append(networks, parsed...)
		seen[base] = true
	}
	if !seen["GeoLite2-ASN-Blocks-IPv4.csv"] || !seen["GeoLite2-ASN-Blocks-IPv6.csv"] {
		return nil, errors.New("archive missing required IPv4 or IPv6 ASN block file")
	}
	if len(networks) == 0 {
		return nil, errors.New("archive contains no ASN block files")
	}
	return networks, nil
}

func parseASNCSV(r io.Reader) ([]asnNetwork, error) {
	records := csv.NewReader(r)
	records.ReuseRecord = true

	header, err := records.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}
	if len(header) != 3 || header[0] != "network" || header[1] != "autonomous_system_number" || header[2] != "autonomous_system_organization" {
		return nil, errors.New("invalid ASN CSV header")
	}

	var networks []asnNetwork
	for {
		record, err := records.Read()
		if errors.Is(err, io.EOF) {
			return networks, nil
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV row: %w", err)
		}

		if len(record) != 3 {
			return nil, errors.New("invalid ASN CSV row")
		}
		prefix, err := netip.ParsePrefix(record[0])
		if err != nil {
			return nil, fmt.Errorf("invalid network %q", record[0])
		}
		asn, err := strconv.ParseUint(record[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid ASN %q", record[1])
		}
		if len(record[2]) > 4096 {
			return nil, errors.New("organization field too large")
		}

		networks = append(networks, asnNetwork{
			network:      prefix.Masked(),
			asn:          int64(asn),
			organization: record[2],
		})
	}
}
