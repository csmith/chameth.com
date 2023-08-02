import md from "../libraries/markdown.js";
import liquid from "../libraries/liquid.js";

export default async (page) => {
    return md.render(
        await liquid.parseAndRender(page.excerpt, {page})
    ).replaceAll(/\[\^\d+]/g, '');
}