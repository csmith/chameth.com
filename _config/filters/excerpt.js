import md from "../libraries/markdown.js";
import liquid from "../libraries/liquid.js";
import {readFrontMatter} from "../_lib/frontmatter.js";

export default async (page) => {
    if (page.inputPath && page.inputPath.includes('/posts')) {
        const frontMatter = readFrontMatter(page.inputPath);
        if (frontMatter.format !== 'short' && frontMatter.format !== 'long') {
            throw new Error(`Unknown format in post "${page.inputPath}"`);
        }

        if (!page.excerpt && frontMatter.format === 'long') {
            throw new Error(`Missing <!--more--> in long-format post "${page.inputPath}"`);
        }

        if (frontMatter.format === 'short' && !page.excerpt) {
            return md.render(await liquid.parseAndRender(page.rawInput, {page})).replaceAll(/\[\^\d+]/g, '')
        }
    }

    return page.excerpt && md.render(
        await liquid.parseAndRender(page.excerpt, {page})
    ).replaceAll(/\[\^\d+]/g, '') || '';
}