import {readFrontMatter} from "../../_lib/frontmatter.js";
import njk from "../../libraries/nunjucks.js";

export default function (caption) {
    const data = readFrontMatter(this.page.inputPath)
    const resource = data.resources.find((r) => r.name === caption)

    return njk.renderString(
        '<figure class="full">' +
        '    <video src="{{ src }}" alt="{{ name }}" controls></video>' +
        '</figure>',
        {
            src: resource.src,
            name: resource.name,
        },
    );
}