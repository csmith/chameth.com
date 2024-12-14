import isoDateFormat from "./isodateformat.js";
import dateFormat from "./dateformat.js";
import excerpt from "./excerpt.js";
import head from "./head.js";

export default function (eleventyConfig) {
    eleventyConfig.addFilter("isoDateFormat", isoDateFormat);
    eleventyConfig.addFilter("dateFormat", dateFormat);
    eleventyConfig.addFilter("excerpt", excerpt);
    eleventyConfig.addFilter("head", head);
};