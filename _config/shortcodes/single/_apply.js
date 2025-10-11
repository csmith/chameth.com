import audio from "./audio.js";
import figure from "./figure.js";
import img from "./img.js";
import video from "./video.js";

export default function (eleventyConfig) {
    eleventyConfig.addShortcode("audio", audio);
    eleventyConfig.addShortcode("figure", figure);
    eleventyConfig.addShortcode("img", img);
    eleventyConfig.addShortcode("video", video);
}