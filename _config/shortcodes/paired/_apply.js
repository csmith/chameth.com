import poem from './poem.js'
import sidenote from './sidenote.js'
import update from "./update.js";
import warning from "./warning.js";

export default function (eleventyConfig) {
    eleventyConfig.addPairedShortcode("poem", poem);
    eleventyConfig.addPairedShortcode("sidenote", sidenote);
    eleventyConfig.addPairedShortcode("update", update);
    eleventyConfig.addPairedShortcode("warning", warning);
}