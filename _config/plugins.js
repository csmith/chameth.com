import rev from "eleventy-plugin-rev"
import rss from "@11ty/eleventy-plugin-rss"
import sass from "eleventy-sass"

export default function (eleventyConfig) {
    eleventyConfig.addPlugin(rev);
    eleventyConfig.addPlugin(rss);
    eleventyConfig.addPlugin(sass, {rev: true});
}