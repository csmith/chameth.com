import applySingle from './single/_apply.js'
import applyPaired from './paired/_apply.js'

export default function (eleventyConfig) {
    applySingle(eleventyConfig);
    applyPaired(eleventyConfig);
}