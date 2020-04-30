<template>
    <div
            style="height: 100%;
    background: #002833;"
    >
        <div id="terminal" ref="terminal"></div>
    </div>
</template>
<script>
    import "xterm/dist/xterm.css";
    import { Terminal } from "xterm";
    import * as fit from "xterm/lib/addons/fit/fit";
    import {getQueryVariable} from '@/utils/tools'
    import { WSAddr } from '@/axios/job';

    Terminal.applyAddon(fit); // Apply the `fit` addon

    let defaultTheme = {
        foreground: "#ffffff",
        background: "#1b212f",
        cursor: "#ffffff",
        selection: "rgba(255, 255, 255, 0.3)",
        black: "#000000",
        brightBlack: "#808080",
        red: "#ce2f2b",
        brightRed: "#f44a47",
        green: "#00b976",
        brightGreen: "#05d289",
        yellow: "#e0d500",
        brightYellow: "#f4f628",
        magenta: "#bd37bc",
        brightMagenta: "#d86cd8",
        blue: "#1d6fca",
        brightBlue: "#358bed",
        cyan: "#00a8cf",
        brightCyan: "#19b8dd",
        white: "#e5e5e5",
        brightWhite: "#ffffff"
    };
    let bindTerminalResize = (term, websocket) => {
        let onTermResize = size => {
            websocket.send(
                JSON.stringify({
                    type: "resize",
                    rows: size.rows,
                    cols: size.cols
                })
            );
        };
        // register resize event.
        term.on("resize", onTermResize);
        // unregister resize event when WebSocket closed.
        websocket.addEventListener("close", function () {
            term.off("resize", onTermResize);
        });
    };
    let bindTerminal = (term, websocket, bidirectional, bufferedTime) => {
        term.socket = websocket;
        let messageBuffer = null;
        let handleWebSocketMessage = function (ev) {
            if (bufferedTime && bufferedTime > 0) {
                if (messageBuffer) {
                    messageBuffer += ev.data;
                } else {
                    messageBuffer = ev.data;
                    setTimeout(function () {
                        term.write(messageBuffer);
                    }, bufferedTime);
                }
            } else {
                term.write(ev.data);
            }
        };
        let handleTerminalData = function (data) {
            websocket.send(
                JSON.stringify({
                    type: "input",
                    input: data // encode data as base64 format
                })
            );
        };
        websocket.onmessage = handleWebSocketMessage;
        if (bidirectional) {
            term.on("data", handleTerminalData);
        }

        // send heartbeat package to avoid closing webSocket connection in some proxy environmental such as nginx.
        let heartBeatTimer = setInterval(function () {
            websocket.send(JSON.stringify({type: "heartbeat", data: ""}));
        }, 20 * 1000);

        websocket.addEventListener("close", function () {
            websocket.removeEventListener("message", handleWebSocketMessage);
            term.off("data", handleTerminalData);
            delete term.socket;
            clearInterval(heartBeatTimer);
        });
    };
    export default {
        data() {
            return {
                copy: ""
            };
        },
        method: {
            onWindowResize() {
                this.term.fit(); // it will make terminal resized.
            },

        },
        mounted() {
            let podNs = getQueryVariable("ns");
            let podName =  getQueryVariable("name");
            let containerName =  getQueryVariable("cname");
            console.log(podNs,podName,containerName);
            let terminalContainer = this.$refs["terminal"];
            let rows = window.outerHeight / 16 - 13;
            let cols = document.body.offsetWidth / 10;
            let term = new Terminal({
                rows: parseInt(rows),
                cols: parseInt(cols),
                rendererType: "canvas", //渲染类型
                fontSize: 16, //字体大小
                cursorBlink: true, //光标闪烁
                cursorStyle: 'underline',
                bellStyle: "sound",
                theme: defaultTheme
            });
            term.open(terminalContainer, true);
            // 当浏览器窗口变化时, 重新适配终端
            window.addEventListener("resize",this.onWindowResize);
            term.fit();
            let ws_url = `${WSAddr}/ssh?podNs=${podNs}&podName=${podName}&containerName=${containerName}`;
            let websocket = new WebSocket(ws_url);//地址
            //连接成功
            websocket.onopen = function(evt) {
                console.log("onopen", evt);
            };
            //关闭
            websocket.onclose = function(evt) {
                console.log("close", evt);
            };
            //错误
            websocket.onerror = function(evt) {
                console.log("error", evt);
            };
            bindTerminal(term, websocket, true, -1);
            bindTerminalResize(term, websocket);
        }
    };
</script>

<style scoped>
</style>
