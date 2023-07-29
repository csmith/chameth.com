const fs = require('fs');
const path = require('path');
const matter = require("gray-matter");
const yaml = require("js-yaml");
const markdownIt = require('markdown-it');
const {Liquid} = require("liquidjs");
const Nunjucks = require('nunjucks');
const sizeOf = require('image-size');
const {DateTime} = require("luxon");
const prettier = require('prettier');

const md = markdownIt({
    html: true,
    typographer: true,
})
    .use(require('markdown-it-footnote'))
    .use(require('markdown-it-prism'));

md.renderer.rules.table_open = () => '<div class="table-holder"><table>';
md.renderer.rules.table_close = () => '</table></div>';

const liquid = new Liquid({
    extname: ".liquid",
    dynamicPartials: false,
    strictFilters: true,
    root: ["_includes"]
});

const njk = new Nunjucks.Environment(
    new Nunjucks.FileSystemLoader("_includes")
);

const readFrontMatter = (f) => matter(fs.readFileSync(f)).data;

const renamer = function (original) {
    // See if there's an index.md alongside the asset, if so figure out the permalink and put the file there.
    // This emulates Hugo's "page bundles" behaviour.
    const content = path.join(path.dirname(original), 'index.md');
    if (fs.existsSync(content)) {
        const data = readFrontMatter(content);
        if (data && data.permalink) {
            return path.join(data.permalink, path.basename(original));
        }
    }

    // Dump content in /static at the root of the output.
    if (original.startsWith('static/')) {
        return original.replace('static/', '');
    }
    return original;
};

const poem = function (content) {
    return njk.renderString(
        '<section class="poem">\n{{ content | safe }}\n</section>',
        {
            content: content.trim().replaceAll('\n', '<br>'),
        },
    );
};

const update = function (content, date) {
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

const sidenote = function (content, title) {
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

const video = function (caption) {
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

const img = function (caption) {
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

const figure = function (_class, caption) {
    const data = readFrontMatter(this.page.inputPath)
    const resource = data.resources.find((r) => r.name === caption)
    const size = sizeOf(path.join(path.dirname(this.page.inputPath), resource.src));
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

const head = (array, n) => {
    if (!Array.isArray(array) || array.length === 0) {
        return [];
    }
    if (n < 0) {
        return array.slice(n);
    }

    return array.slice(0, n);
};

const dateFormat = (date) => DateTime.fromJSDate(date).toLocaleString(DateTime.DATE_MED);

const excerpt = async (page) => {
    return md.render(await liquid.parseAndRender(page.excerpt, {page})).replaceAll(/\[\^\d+]/g, '');
}

const tidyHTML = function (content) {
    if (this.page.outputPath && this.page.outputPath.endsWith(".html") && content) {
        return prettier.format(content, {
            parser: "html",
            printWidth: 120,
            bracketSameLine: true,
        });
    } else {
        return content;
    }
}

const tidyCSS = function (content) {
    if (this.page.outputPath && this.page.outputPath.endsWith(".css") && content) {
        return prettier.format(content, {
            parser: "css",
            printWidth: 120,
            bracketSameLine: true,
        });
    } else {
        return content;
    }
}

module.exports = {
    dataExtensions: {
        "yml": (contents) => yaml.load(contents),
    },
    plugins: [
        require("eleventy-plugin-rev"),
        require("@11ty/eleventy-plugin-rss"),
        [
            require("eleventy-sass"),
            {
                rev: true,
            }
        ],
    ],
    libraries: {
        md,
        liquid,
        njk,
    },
    assets: {
        filter: [
            "content/**/*.jpg",
            "content/**/*.webp",
            "content/**/*.webm",
            "content/**/*.avif",
            "content/**/*.png",
            "static/*.*",
        ],
        renamer,
    },
    filters: {
        head,
        excerpt,
        dateFormat,
    },
    shortcodes: {
        single: {
            video,
            figure,
            img,
        },
        paired: {
            poem,
            sidenote,
            update,
        },
    },
    transforms: {
        tidyHTML,
        tidyCSS,
    }
}