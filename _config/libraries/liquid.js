import {Liquid} from 'liquidjs';

const liquid = new Liquid({
    extname: ".liquid",
    dynamicPartials: false,
    strictFilters: true,
    root: ["_includes"]
});

export default liquid;