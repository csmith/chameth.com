import longposts from "./longposts.js";
import shortposts from "./shortposts.js";
import sortedsnippets from "./sortedsnippets.js";

export default function (eleventyConfig) {
    eleventyConfig.addCollection("longPosts", longposts);
    eleventyConfig.addCollection("shortPosts", shortposts);
    eleventyConfig.addCollection("sortedSnippets", sortedsnippets);
};