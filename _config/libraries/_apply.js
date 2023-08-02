import md from "./markdown.js";
import liquid from "./liquid.js";
import njk from "./nunjucks.js";

export default function (eleventyConfig) {
    eleventyConfig.setLibrary("md", md);
    eleventyConfig.setLibrary("liquid", liquid);
    eleventyConfig.setLibrary("njk", njk);
}