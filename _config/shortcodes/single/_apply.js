import figure from "./figure.js";
import img from "./img.js";
import stylesheet from "./stylesheet.js";
import video from "./video.js";

export default function (eleventyConfig) {
    eleventyConfig.addShortcode("figure", figure);
    eleventyConfig.addShortcode("img", img);
    eleventyConfig.addShortcode("stylesheet", stylesheet);
    eleventyConfig.addShortcode("video", video);
}