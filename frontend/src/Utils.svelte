<script context="module">
    /*
     * Copyright (C) 2024- Germano Rizzo
     *
     * This file is part of Seif.
     *
     * Seif is free software: you can redistribute it and/or modify
     * it under the terms of the GNU General Public License as published by
     * the Free Software Foundation, either version 3 of the License, or
     * (at your option) any later version.
     *
     * Seif is distributed in the hope that it will be useful,
     * but WITHOUT ANY WARRANTY; without even the implied warranty of
     * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
     * GNU General Public License for more details.
     *
     * You should have received a copy of the GNU General Public License
     * along with Seif.  If not, see <http://www.gnu.org/licenses/>.
     */

    // Common routines

    // REST stuff

    function mapToUrl(map) {
        let first = true;
        let urlPiece = "";
        for (const [key, value] of Object.entries(map)) {
            if (first) {
                urlPiece += "?";
                first = false;
            } else {
                urlPiece += "&";
            }
            urlPiece += key + "=" + encodeURI(value);
        }
        return urlPiece;
    }

    // @ts-ignore
    export const CALL = async function (
        srv,
        method = "GET",
        json = null,
        map = null,
        timeout = 5000,
    ) {
        let url = "/api/" + srv;
        if (!!map) url += mapToUrl(map);

        const req = {
            method: method,
            signal: AbortSignal.timeout(timeout),
        };
        if (method === "PUT" || method === "POST") {
            req["body"] = !!json ? JSON.stringify(json) : "{}";
            req["headers"] = { "Content-Type": "application/json" };
        }

        try {
            const res = await fetch(url, req);

            const ret = {
                isErr: !res.ok,
                status: res.status,
            };

            if (res.headers.get("Content-Type") == "application/json") {
                if (res.ok) ret.payload = await res.json();
                else {
                    const err = await res.json();
                    let msg = err.code;
                    msg = msg.charAt(0).toUpperCase() + msg.slice(1);
                    if (msg.includes("%s")) msg = msg.replace("%s", err.object);
                    if (!!err.error)
                        console.error("!!ERROR!!" + msg + ": " + err.error);
                    ret.message = msg;
                }
            } else ret.message = await res.text();

            return ret;
        } catch (e) {
            return {
                isErr: true,
                status: 599,
                message: e,
            };
        }
    };

    import CryptoJS from "crypto-js";

    export const GEN_PASSWORD = function () {
        return CryptoJS.lib.WordArray.random(128 / 8);
    };

    export const PASS2STR = function (pass, removeEquals = false) {
        if (removeEquals)
            return CryptoJS.enc.Base64.stringify(pass).replaceAll("=", "");
        return CryptoJS.enc.Base64.stringify(pass);
    };

    export const STR2PASS = function (pass, addEquals = 0) {
        for (let i = 0; i < addEquals; i++) pass += "=";
        return CryptoJS.enc.Base64.parse(pass);
    };

    export const ENCRYPT = function (message) {
        const iv = GEN_PASSWORD();
        const key = GEN_PASSWORD();

        const encrypted = CryptoJS.AES.encrypt(message, key, {
            iv,
            padding: CryptoJS.pad.Pkcs7,
            mode: CryptoJS.mode.CTR,
        });

        const sha = CryptoJS.SHA256(message);

        return {
            key: PASS2STR(key, true),
            iv: PASS2STR(iv, true),
            text: encrypted.toString(),
            sha: PASS2STR(sha, true),
        };
    };

    export const DECRYPT = function (encoded) {
        const key = STR2PASS(encoded.key, 2);
        const iv = STR2PASS(encoded.iv, 2);

        let decrypted = "";
        try {
            decrypted = CryptoJS.AES.decrypt(encoded.text, key, {
                iv,
                padding: CryptoJS.pad.Pkcs7,
                mode: CryptoJS.mode.CTR,
            });
        } catch (e) {
            throw e + ". Wrong password?";
        }

        const ret = decrypted.toString(CryptoJS.enc.Utf8);

        const sha = CryptoJS.SHA256(ret);

        if (PASS2STR(sha, true) != encoded.sha)
            throw "Checksum failed. Wrong password?";

        return decrypted.toString(CryptoJS.enc.Utf8);
    };

    import Swal from "sweetalert2";

    export const TOAST = async function (message) {
        await Swal.fire({
            text: message,
            toast: true,
            position: "top",
            timer: 1000,
            showConfirmButton: false,
        });
    };

    export const ERROR = async function (message) {
        await Swal.fire({
            title: "Error!",
            text: message,
            icon: "error",
            position: "top",
            toast: true,
            confirmButtonText: "Ouch!",
        });
    };
</script>
