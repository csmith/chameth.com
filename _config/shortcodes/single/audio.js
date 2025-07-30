import {readFrontMatter} from "../../_lib/frontmatter.js";
import njk from "../../libraries/nunjucks.js";

export default function (caption) {
    const data = readFrontMatter(this.page.inputPath)
    const resource = data.resources.find((r) => r.name === caption)

    return njk.renderString(
        '<figure class="full">' +
        '    <audio src="{{ src }}" alt="{{ name }}" controls></audio>' +
        '    <figcaption>{{ caption }}</figcaption>' +
        '</figure>',
        {
            src: resource.src,
            name: resource.name,
            caption: resource.title ?? resource.name,
        },
    );
}