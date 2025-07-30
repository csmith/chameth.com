import njk from "../../libraries/nunjucks.js";
import md from "../../libraries/markdown.js";

export default function (content) {
    return njk.renderString(
        '<aside class="warning">' +
        '    <div>' +
        '        <h5>Warning</h5>' +
        '        {{ content | safe }}' +
        '    </div>' +
        '</aside>',
        {
            content: md.render(content.trim()),
        },
    );
};