import prettier from "prettier";

export default function (content) {
    if (this.page.outputPath && this.page.outputPath.endsWith(".css") && content) {
        return prettier.format(content, {
            parser: "css",
            printWidth: 120,
            bracketSameLine: true,
        });
    } else {
        return content;
    }
}