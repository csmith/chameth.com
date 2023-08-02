import sortedsnippets from "./sortedsnippets.js";

export default function (eleventyConfig) {
    eleventyConfig.addCollection("sortedSnippets", sortedsnippets);
};