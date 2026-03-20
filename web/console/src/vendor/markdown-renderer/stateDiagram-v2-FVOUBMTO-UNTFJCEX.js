import {
  StateDB,
  stateDiagram_default,
  stateRenderer_v3_unified_default,
  styles_default
} from "./chunk-XXFTOW7Q.js";
import "./chunk-FIQB4NCH.js";
import "./chunk-EIDMTQSE.js";
import "./chunk-PKGDMGHH.js";
import "./chunk-ZSAKS3XY.js";
import "./chunk-MOB7WDQN.js";
import "./chunk-ES3KF662.js";
import "./chunk-XRTBQEWT.js";
import "./chunk-SD42X4BU.js";
import "./chunk-BBND2PSY.js";
import "./chunk-WRXS75UJ.js";
import "./chunk-2F5OFAT2.js";
import "./chunk-CYX77XRJ.js";
import "./chunk-4GOL3AE5.js";
import "./chunk-NV3YER2Z.js";
import {
  __name
} from "./chunk-CMV2AIJA.js";
import "./chunk-J36ZT73G.js";
import "./chunk-TBFEXUVH.js";
import "./chunk-ORRSYHNQ.js";

// node_modules/.pnpm/mermaid@11.13.0/node_modules/mermaid/dist/chunks/mermaid.core/stateDiagram-v2-FVOUBMTO.mjs
var diagram = {
  parser: stateDiagram_default,
  get db() {
    return new StateDB(2);
  },
  renderer: stateRenderer_v3_unified_default,
  styles: styles_default,
  init: /* @__PURE__ */ __name((cnf) => {
    if (!cnf.state) {
      cnf.state = {};
    }
    cnf.state.arrowMarkerAbsolute = cnf.arrowMarkerAbsolute;
  }, "init")
};
export {
  diagram
};
