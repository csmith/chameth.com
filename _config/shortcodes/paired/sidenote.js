import njk from "../../libraries/nunjucks.js";
import md from "../../libraries/markdown.js";

export default function (content, title) {
    return njk.renderString(
        '<aside class="sidenote">' +
        '    <h5>Side note: {{ title }}</h5>' +
        '    {{ content | safe }}' +
        '</aside>',
        {
            title,
            content: md.render(content.trim()),
        },
    );
};