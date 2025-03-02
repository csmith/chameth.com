import { imageSizeFromFile } from 'image-size/fromFile';
import path from 'path';
import njk from "../../libraries/nunjucks.js";
import {readFrontMatter} from "../../_lib/frontmatter.js";

export default async function (_class, caption) {
    const data = readFrontMatter(this.page.inputPath)
    const resource = data.resources.find((r) => r.name === caption)
    const size = await imageSizeFromFile(path.join(path.dirname(this.page.inputPath), resource.src));
    const baseName = resource.src.replace(/\.(jpg|png)$/, '');

    return njk.renderString(
        '<figure class="{{ _class }}">' +
        '    <picture>' +
        '        <source srcset="{{ prefix }}{{ baseName }}.avif" type="image/avif">' +
        '        <source srcset="{{ prefix }}{{ baseName }}.webp" type="image/webp">' +
        '        <img src="{{ prefix }}{{ src }}" alt="{{ name }}" loading="lazy" width="{{ size.width }}" height="{{ size.height }}">' +
        '    </picture>' +
        '    <figcaption>{{ caption }}</figcaption>' +
        '</figure>',
        {
            _class,
            prefix: this.page.url,
            baseName,
            src: resource.src,
            name: resource.name,
            caption: resource.title ?? resource.name,
            size,
        },
    );
};