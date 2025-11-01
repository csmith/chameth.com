import {binaryInlined, Dialect, WorkerLinter} from '/assets/harper/harper.js';

// Initialize Harper linter once for all textareas
const linter = new WorkerLinter({
    binary: binaryInlined,
    dialect: Dialect.British
});

// Helper function to escape HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Find all markdown syntax elements
function findMarkdownSyntax(text) {
    const syntaxSpans = [];

    // Fenced code blocks
    const codeBlockRegex = /^```[\s\S]*?^```/gm;
    let match;
    while ((match = codeBlockRegex.exec(text)) !== null) {
        syntaxSpans.push({
            start: match.index,
            end: match.index + match[0].length,
            type: 'codeblock',
            color: '#586e75'
        });
    }

    // Headers
    const headerRegex = /^#{3,6} .*/gm;
    while ((match = headerRegex.exec(text)) !== null) {
        syntaxSpans.push({
            start: match.index,
            end: match.index + match[0].length,
            type: 'header',
            color: '#0066cc'
        });
    }

    // Inline links
    const inlineLinkRegex = /\[([^\]]+)]\(([^)]+)\)/g;
    while ((match = inlineLinkRegex.exec(text)) !== null) {
        syntaxSpans.push({
            start: match.index,
            end: match.index + match[0].length,
            type: 'link',
            color: '#0066cc'
        });
    }

    // Footnotes [^...]
    const footnoteRegex = /\[\^[^\]]+]/g;
    while ((match = footnoteRegex.exec(text)) !== null) {
        syntaxSpans.push({
            start: match.index,
            end: match.index + match[0].length,
            type: 'footnote',
            color: '#6c71c4'
        });
    }

    // Shortcodes {% ... %}
    const shortcodeRegex = /\{%[^%]*%}/g;
    while ((match = shortcodeRegex.exec(text)) !== null) {
        syntaxSpans.push({
            start: match.index,
            end: match.index + match[0].length,
            type: 'shortcode',
            color: '#d33682'
        });
    }

    return syntaxSpans;
}

// Build highlighted text with both syntax highlighting and error spans
function buildHighlightedText(text, lints) {
    const syntaxSpans = findMarkdownSyntax(text);

    // Convert lints to spans
    const errorSpans = lints.map(lint => ({
        start: lint.span().start,
        end: lint.span().end,
        type: 'error',
        style: 'background-color: rgba(255, 140, 0, 0.08); border-bottom: 1px solid #ff8c00;'
    }));

    // Merge all spans
    const allSpans = [...syntaxSpans, ...errorSpans];

    // Sort by start position
    allSpans.sort((a, b) => a.start - b.start);

    let result = '';
    let lastIndex = 0;

    for (const span of allSpans) {
        // Add text before this span
        if (span.start > lastIndex) {
            result += escapeHtml(text.substring(lastIndex, span.start));
        }

        // Don't process overlapping spans (skip if we already processed this text)
        if (span.start < lastIndex) {
            continue;
        }

        // Add highlighted span
        const spanText = text.substring(span.start, span.end);
        if (span.type === 'error') {
            result += `<span style="${span.style}">${escapeHtml(spanText)}</span>`;
        } else {
            result += `<span style="color: ${span.color};">${escapeHtml(spanText)}</span>`;
        }

        lastIndex = span.end;
    }

    // Add remaining text
    if (lastIndex < text.length) {
        result += escapeHtml(text.substring(lastIndex));
    }

    return result;
}

// Enhance a single textarea with Harper and syntax highlighting
function enhanceTextarea(textarea) {
    // Skip if already enhanced
    if (textarea.dataset.enhanced === 'true') {
        return;
    }
    textarea.dataset.enhanced = 'true';

    // Create wrapper
    const wrapper = document.createElement('div');
    wrapper.style.cssText = 'position: relative; width: 100%; background: #fff;';

    // Create mirror div
    const mirror = document.createElement('div');
    mirror.className = 'textarea-mirror';
    mirror.style.cssText = `
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        padding: 0.5rem;
        border: 1px solid transparent;
        white-space: pre-wrap;
        word-wrap: break-word;
        overflow-wrap: break-word;
        overflow: auto;
        pointer-events: none;
        font-family: 'Berkeley Mono', monospace;
        font-size: 1.25rem;
        line-height: 1.5;
        color: #000;
        background: #fff;
        z-index: 1;
    `;

    // Create tooltip
    const tooltip = document.createElement('div');
    tooltip.className = 'textarea-tooltip';
    tooltip.style.cssText = `
        position: absolute;
        display: none;
        background: #fff;
        color: #333;
        padding: 0.5rem 0.75rem;
        border-radius: 4px;
        font-size: 0.875rem;
        max-width: 300px;
        z-index: 10;
        pointer-events: none;
        box-shadow: 0 2px 8px rgba(0,0,0,0.3);
        border: 1px solid #ccc;
        white-space: normal;
        word-wrap: break-word;
    `;

    // Wrap textarea
    textarea.parentNode.insertBefore(wrapper, textarea);
    wrapper.appendChild(mirror);
    wrapper.appendChild(textarea);
    wrapper.appendChild(tooltip);

    // Style textarea
    textarea.spellcheck = false;
    textarea.style.cssText = `
        position: relative;
        width: 100%;
        height: 400px;
        padding: 0.5rem;
        font-family: 'Berkeley Mono', monospace;
        font-size: 1.25rem;
        line-height: 1.5;
        color: transparent;
        caret-color: #000;
        background: transparent;
        z-index: 2;
        resize: vertical;
    `;

    let currentLints = [];

    // Sync scroll position from textarea to mirror
    function syncScroll() {
        mirror.scrollTop = textarea.scrollTop;
        mirror.scrollLeft = textarea.scrollLeft;
        // Hide tooltip when scrolling
        tooltip.style.display = 'none';
    }

    // Get approximate pixel coordinates of a text position in the textarea
    function getTextPosition(textPos) {
        const text = textarea.value.substring(0, textPos);
        const lines = text.split('\n');
        const lineNumber = lines.length - 1;

        // Get computed styles
        const styles = window.getComputedStyle(textarea);
        const lineHeight = parseFloat(styles.lineHeight);
        const paddingTop = parseFloat(styles.paddingTop);
        const paddingLeft = parseFloat(styles.paddingLeft);

        // Calculate approximate position
        const y = paddingTop + (lineNumber * lineHeight) - textarea.scrollTop;
        const x = paddingLeft; // Simplified - just use left edge

        return { x, y, lineHeight };
    }

    // Check if cursor is within an error and show tooltip
    function checkCursorPosition() {
        const cursorPos = textarea.selectionStart;

        // Find lint at cursor position
        const lintAtCursor = currentLints.find(lint => {
            const span = lint.span();
            return cursorPos >= span.start && cursorPos <= span.end;
        });

        if (lintAtCursor) {
            // Show tooltip near cursor
            tooltip.textContent = lintAtCursor.message();
            tooltip.style.display = 'block';

            // Position tooltip near the error
            const pos = getTextPosition(cursorPos);
            const textareaRect = textarea.getBoundingClientRect();

            // Position to the right of the cursor, or below if too wide
            tooltip.style.left = Math.min(pos.x + 200, textareaRect.width - 320) + 'px';
            tooltip.style.top = (pos.y + pos.lineHeight) + 'px';

            // Make sure tooltip doesn't go off the bottom
            const tooltipRect = tooltip.getBoundingClientRect();
            if (tooltipRect.bottom > textareaRect.bottom) {
                tooltip.style.top = (pos.y - tooltipRect.height - 5) + 'px';
            }
        } else {
            tooltip.style.display = 'none';
        }
    }

    // Update mirror immediately with plain text, then with syntax highlighting
    function updateMirrorImmediate() {
        const text = textarea.value;
        // Show plain text immediately
        mirror.innerHTML = escapeHtml(text);
        // Then update with syntax highlighting and linting
        updateMirrorAsync();
    }

    // Update mirror with highlighted content (async)
    let lintInProgress = false;
    async function updateMirrorAsync() {
        if (lintInProgress) return;
        lintInProgress = true;

        const text = textarea.value;

        try {
            const lints = await linter.lint(text);
            // Check if text hasn't changed while linting
            if (text === textarea.value) {
                currentLints = lints;
                mirror.innerHTML = buildHighlightedText(text, lints);
                checkCursorPosition();
            }
        } catch (error) {
            console.error('Harper linting error:', error);
            currentLints = [];
            mirror.innerHTML = escapeHtml(text);
        } finally {
            lintInProgress = false;
        }
    }

    // Add event listeners
    textarea.addEventListener('input', updateMirrorImmediate);
    textarea.addEventListener('scroll', syncScroll);
    textarea.addEventListener('click', checkCursorPosition);
    textarea.addEventListener('keyup', checkCursorPosition);
    textarea.addEventListener('select', checkCursorPosition);

    // Handle textarea resize to update mirror height
    const resizeObserver = new ResizeObserver(() => {
        syncScroll();
    });
    resizeObserver.observe(textarea);

    // Run initial update - show text immediately, then add syntax highlighting
    updateMirrorImmediate();
}

document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('textarea').forEach(enhanceTextarea);
});
