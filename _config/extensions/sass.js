import path from "path";
import * as sass from "sass";
import {sassTarget} from "../_lib/sass.js";

const compiler = sass.initCompiler();

export default {
    outputFileExtension: 'css',

    compileOptions: {
        cache: false,

        permalink: (inputContent, inputPath) => {
            if (path.parse(inputPath).name.startsWith('_')) {
                return false;
            } else {
                return sassTarget();
            }
        }
    },

    compile: (inputContent, inputPath) => {
        const {css} = compiler.compileString(inputContent, {
            loadPaths: [path.parse(inputPath).dir || '.', 'node_modules']
        });

        return () => css;
    }
}