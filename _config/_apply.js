import applyFrontmatter from './frontmatter.js'
import applyPlugins from './plugins.js'
import applyAssets from './assets/_apply.js'
import applyCollections from './collections/_apply.js'
import applyDataExtensions from './dataextensions/_apply.js'
import applyExtensions from './extensions/_apply.js'
import applyFilters from './filters/_apply.js'
import applyLibraries from './libraries/_apply.js'
import applyShortcodes from './shortcodes/_apply.js'
import applyTransforms from './transforms/_apply.js'
import applyWatches from './watches/_apply.js'

export default function (eleventyConfig) {
    applyLibraries(eleventyConfig);
    applyPlugins(eleventyConfig);

    applyAssets(eleventyConfig);
    applyCollections(eleventyConfig);
    applyDataExtensions(eleventyConfig);
    applyExtensions(eleventyConfig);
    applyFilters(eleventyConfig);
    applyFrontmatter(eleventyConfig);
    applyShortcodes(eleventyConfig);
    applyTransforms(eleventyConfig);
    applyWatches(eleventyConfig);
}