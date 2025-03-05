export default function (eleventyConfig) {
    eleventyConfig.addWatchTarget("./static/**/*.scss", {
        resetConfig: true
    });

    eleventyConfig.addWatchTarget("./_config/**/*.js", {
        resetConfig: true
    });
};