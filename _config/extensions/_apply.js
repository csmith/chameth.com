import sass from "./sass.js";

export default function (eleventyConfig) {
    eleventyConfig.addTemplateFormats("scss");
    eleventyConfig.addExtension("scss", sass);
};