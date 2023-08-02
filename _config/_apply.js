import applyFrontmatter from './frontmatter.js'
import applyPlugins from './plugins.js'
import applyAssets from './assets/_apply.js'
import applyCollections from './collections/_apply.js'
import applyFilters from './filters/_apply.js'
import applyLibraries from './libraries/_apply.js'
import applyShortcodes from './shortcodes/_apply.js'
import applyTransforms from './transforms/_apply.js'

export default function (eleventyConfig) {
    applyLibraries(eleventyConfig);
    applyPlugins(eleventyConfig);

    applyFrontmatter(eleventyConfig);
    applyAssets(eleventyConfig);
    applyCollections(eleventyConfig);
    applyFilters(eleventyConfig);
    applyShortcodes(eleventyConfig);
    applyTransforms(eleventyConfig);
}