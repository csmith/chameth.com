import filter from "./filter.js";
import rename from "./renamer.js";

export default function (eleventyConfig) {
    eleventyConfig.addPassthroughCopy(".", {
        filter,
        rename,
    });
}