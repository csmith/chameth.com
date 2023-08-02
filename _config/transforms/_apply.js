import tidyCSS from "./tidycss.js";
import tidyHTML from "./tidyhtml.js";

export default function (eleventyConfig) {
    eleventyConfig.addTransform("tidyCSS", tidyCSS);
    eleventyConfig.addTransform("tidyHTML", tidyHTML);
};