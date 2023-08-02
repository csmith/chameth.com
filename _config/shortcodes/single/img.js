import sizeOf from 'image-size';
import path from 'path';
import njk from "../../libraries/nunjucks.js";
import {readFrontMatter} from "../../_lib/frontmatter.js";


export default function (caption) {
    const data = readFrontMatter(this.page.inputPath)
    const resource = data.resources.find((r) => r.name === caption)
    const size = sizeOf(path.join(path.dirname(this.page.inputPath), resource.src));
    const baseName = resource.src.replace(/\.(jpg|png)$/, '');

    return njk.renderString(
        '<figure class="full">' +
        '    <picture>' +
        '        <source srcset="{{ baseName }}.avif" type="image/avif">' +
        '        <source srcset="{{ baseName }}.webp" type="image/webp">' +
        '        <img src="{{ src }}" alt="{{ name }}" loading="lazy" width="{{ size.width }}" height="{{ size.height }}">' +
        '    </picture>' +
        '</figure>',
        {
            baseName,
            src: resource.src,
            name: resource.name,
            size,
        },
    );
};