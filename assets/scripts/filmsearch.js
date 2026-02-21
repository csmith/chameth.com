document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll("*[data-form-search]").forEach((el) => {
    el.innerHTML = `
            <div class="film-search">
                <input type="text" name="q" id="q" placeholder="Search films..." autocomplete="off" />
                <div class="results" style="display: none;"></div>
            </div>
        `;

    const input = el.querySelector("input");
    const resultsDiv = el.querySelector(".results");
    let debounceTimer;

    input.addEventListener("input", (e) => {
      const query = e.target.value.trim();

      clearTimeout(debounceTimer);

      if (query.length < 2) {
        resultsDiv.style.display = "none";
        resultsDiv.innerHTML = "";
        return;
      }

      debounceTimer = setTimeout(() => {
        fetch(`/api/films/search?q=${encodeURIComponent(query)}`)
          .then((response) => {
            if (!response.ok) {
              throw new Error("Search failed");
            }
            return response.json();
          })
          .then((films) => {
            if (films.length === 0) {
              resultsDiv.innerHTML = '<p class="no-results">No films found</p>';
              return;
            }

            resultsDiv.innerHTML = films
              .map((film) => {
                const posterSrc =
                  film.poster_path || "/assets/images/default-poster.png";
                const watchText =
                  film.times_watched === 1
                    ? "Watched once"
                    : `Watched ${film.times_watched} times`;

                return `
                        <a href="${film.path}">
                            <img src="${posterSrc}" alt="Poster for ${film.title}" />
                            <div>
                                <h3>${film.title}</h3>
                                <p>${watchText}</p>
                                <p>${film.rating_html}</p>
                            </div>
                        </a>
                    `;
              })
              .join("");
          })
          .finally(() => {
            resultsDiv.style.display = "block";
          })
          .catch(() => {
            resultsDiv.innerHTML =
              '<p class="error">Failed to search films</p>';
          });
      }, 300);
    });

    // Close results when clicking outside
    document.addEventListener("click", (e) => {
      if (!el.contains(e.target)) {
        resultsDiv.style.display = "none";
        resultsDiv.innerHTML = "";
      }
    });
  });
});
