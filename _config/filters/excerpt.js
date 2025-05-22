import md from "../libraries/markdown.js";
import liquid from "../libraries/liquid.js";

export default async (page) => {
    if (!page.excerpt && page.inputPath && page.inputPath.includes('/posts/')) {
        throw new Error(`Missing <!--more--> in post "${page.inputPath}"`);
    }
    
    return page.excerpt && md.render(
        await liquid.parseAndRender(page.excerpt, {page})
    ).replaceAll(/\[\^\d+]/g, '') || '';
}