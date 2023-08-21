import yaml from "./yaml.js";

export default function (eleventyConfig) {
    eleventyConfig.addDataExtension("yml", yaml);
};