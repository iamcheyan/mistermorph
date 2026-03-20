import {
  selectSvgElement
} from "./chunk-QB65KP6J.js";
import {
  parse
} from "./chunk-YYTTR57W.js";
import "./chunk-U46KY2EK.js";
import "./chunk-7JATL2DJ.js";
import "./chunk-UXAP7WXA.js";
import "./chunk-2VQF7DAP.js";
import "./chunk-F7UR3SAG.js";
import "./chunk-KHPNEF2A.js";
import {
  configureSvgSize
} from "./chunk-4GOL3AE5.js";
import "./chunk-NV3YER2Z.js";
import {
  __name,
  log
} from "./chunk-CMV2AIJA.js";
import "./chunk-J36ZT73G.js";
import "./chunk-K6ZM7U6S.js";
import "./chunk-56EAK7OT.js";
import "./chunk-VFJUT2JB.js";
import "./chunk-PQTWO7OT.js";
import "./chunk-TBFEXUVH.js";
import "./chunk-ORRSYHNQ.js";

// node_modules/.pnpm/mermaid@11.13.0/node_modules/mermaid/dist/chunks/mermaid.core/infoDiagram-LFFYTUFH.mjs
var parser = {
  parse: /* @__PURE__ */ __name(async (input) => {
    const ast = await parse("info", input);
    log.debug(ast);
  }, "parse")
};
var DEFAULT_INFO_DB = {
  version: "11.13.0" + (true ? "" : "-tiny")
};
var getVersion = /* @__PURE__ */ __name(() => DEFAULT_INFO_DB.version, "getVersion");
var db = {
  getVersion
};
var draw = /* @__PURE__ */ __name((text, id, version) => {
  log.debug("rendering info diagram\n" + text);
  const svg = selectSvgElement(id);
  configureSvgSize(svg, 100, 400, true);
  const group = svg.append("g");
  group.append("text").attr("x", 100).attr("y", 40).attr("class", "version").attr("font-size", 32).style("text-anchor", "middle").text(`v${version}`);
}, "draw");
var renderer = { draw };
var diagram = {
  parser,
  db,
  renderer
};
export {
  diagram
};
