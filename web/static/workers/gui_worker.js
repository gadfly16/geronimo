var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { guiMessageType } from "../shared/gui_types.js";
// Worker globals
let state = null;
let counter = 0;
// Populate GUI Message handlers
let guiMessageHandlers = {};
guiMessageHandlers[guiMessageType.getUserTree] = getUserStateHandler;
addEventListener("connect", connectHandler);
console.log("Added event listener.");
// Connect to GUI
function connectHandler(e) {
    console.log("Inside web worker.");
    const port = e.ports[0];
    counter += 1;
    port.onmessage = function (e) {
        let gm = e.data;
        guiMessageHandlers[gm.Type](gm)
            .then((resp) => {
            console.log("Worker response:", resp);
            port.postMessage(resp);
        })
            .catch((e) => { console.log(e); });
    };
}
// GUI Message handlers
function getUserStateHandler(msg) {
    return __awaiter(this, void 0, void 0, function* () {
        if (state === null) {
            return fetch("/api/state?" + new URLSearchParams({ userid: msg.Payload }), { method: "get" })
                .then((resp) => resp.json())
                .then((data) => {
                console.log("Data:", data);
                state = data;
                return { Type: guiMessageType.userTree, Payload: data };
            });
        }
        console.log("Returning from state var.");
        return Promise.resolve({ Type: guiMessageType.userTree, Payload: state });
    });
}
