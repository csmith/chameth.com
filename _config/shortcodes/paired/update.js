import njk from "../../libraries/nunjucks.js";
import md from "../../libraries/markdown.js";

export default function (content, date) {
    return njk.renderString(
        '<aside class="update">' +
        '    <h5>Update {{ date }}:</h5>' +
        '    {{ content | safe }}' +
        '</aside>',
        {
            date,
            content: md.render(content.trim()),
        },
    );
};