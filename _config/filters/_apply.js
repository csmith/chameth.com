import isoDateFormat from "./isodateformat.js";
import dateFormat from "./dateformat.js";
import excerpt from "./excerpt.js";
import summary from "./summary.js";
import head from "./head.js";
import related from "./related.js";
import yearsSince from "./yearssince.js";

export default function (eleventyConfig) {
    eleventyConfig.addFilter("isoDateFormat", isoDateFormat);
    eleventyConfig.addFilter("dateFormat", dateFormat);
    eleventyConfig.addFilter("excerpt", excerpt);
    eleventyConfig.addFilter("summary", summary);
    eleventyConfig.addFilter("head", head);
    eleventyConfig.addFilter("related", related);
    eleventyConfig.addFilter("yearsSince", yearsSince);
};