import prettier from "prettier";

export default function (content) {
    if (this.page.outputPath && this.page.outputPath.endsWith(".html") && content) {
        return prettier.format(content, {
            parser: "html",
            printWidth: 120,
            bracketSameLine: true,
        });
    } else {
        return content;
    }
}
