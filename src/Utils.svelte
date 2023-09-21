<script context="module">
    // Common routines

    export const CALL = async function (json) {
        let callObj = {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(json),
        };

        const ret = await fetch("/seif/exec", callObj);

        const retJson = await ret.json();
        retJson.status = ret.status;
        return retJson;
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
        console.error(message);
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
