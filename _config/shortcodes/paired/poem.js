import njk from "../../libraries/nunjucks.js";

export default function (content) {
    return njk.renderString(
        '<section class="poem">\n{{ content | safe }}\n</section>',
        {
            content: content.trim().replaceAll('\n', '<br>'),
        },
    );
};