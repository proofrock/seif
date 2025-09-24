<script>
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
  import ClipboardableField from "./ClipboardableField.svelte";
  import { CALL, ERROR, TOAST } from "./Utils.svelte";
  import { onMount } from "svelte";

  $: initData = null;
  $: token = "";
  $: contents = "";
  $: link = "";
  $: linkNoKey = "";
  $: linkSecret = "";
  $: expiryDays = 3;

  function getParameterByName(name, url = window.location.href) {
    name = name.replace(/[\[\]]/g, "\\$&");
    var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
      results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return "";
    return decodeURIComponent(results[2].replace(/\+/g, " "));
  }

  onMount(async () => {
    const _token = getParameterByName("t");
    token = !_token ? "" : _token;

    const ret = await CALL("getInitData", "GET");
    if (ret.isErr) {
      await ERROR(`Cannot load init data. ${ret.message}.`);
    } else {
      initData = ret.payload;
      expiryDays = initData.default_days;
    }
  });

  async function send() {
    if (contents == "") {
      await ERROR("Empty secret!");
      return;
    }

    if (typeof expiryDays != "number") expiryDays = parseInt(expiryDays);

    if (expiryDays < 1 || isNaN(expiryDays)) {
      await ERROR("Invalid expiration!");
      expiryDays = 7;
      return;
    }

    const obj = {
      secret: contents,
      expiry: expiryDays,
    };
    const ret = await CALL("putSecret", "PUT", obj);
    if (ret.isErr) {
      await ERROR(`Saving failed. ${ret.message}.`);
    } else {
      linkNoKey = `${location.protocol}//${location.host}?t=${ret.payload.id}`;
      link = `${linkNoKey}&s=${encodeURIComponent(ret.payload.key)}`;
      linkSecret = ret.payload.key;
    }
  }

  async function peek() {
    const ret = await CALL("getSecretStatus", "GET", null, { id: token });
    if (ret.isErr) {
      await ERROR(`Status check failed. ${ret.message}.`);
    } else if (ret.payload.pristine) {
      await TOAST("Secret (still) available.");
    } else {
      await TOAST("Secret expired, already revealed or wrong link.");
    }
  }

  async function reveal() {
    let key = getParameterByName("s");
    if (!key) {
      key = prompt("Decoding key").trim();
    }

    const ret = await CALL("getSecret", "DELETE", null, { id: token, key });
    if (ret.isErr) {
      await ERROR(`Secret retrieval failed. ${ret.message}.`);
    } else if (ret.payload.secret === null) {
      await TOAST("Secret expired, already revealed or wrong link.");
    } else {
      contents = ret.payload.secret;
    }
  }
</script>

{#if !!initData}
  <nav class="navbar navbar-expand-lg bg-success text-white">
    <div
      class="container-fluid d-flex justify-content-between align-items-center"
    >
      <div class="navbar-brand bg-success text-white mb-0">
        🔐 Seif
        <span class="small"
          ><small
            ><small>&nbsp;one time secrets drop - {initData.version}</small
            ></small
          ></span
        >
      </div>
      <div>
        <a
          href="https://github.com/proofrock/seif"
          target="_blank"
          style="color: white;"
        >
          <!-- https://github.com/logos -->
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 100 100"
          >
            <path
              fill-rule="evenodd"
              clip-rule="evenodd"
              d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z"
              fill="#fff"
            ></path>
          </svg>
        </a>
      </div>
    </div>
  </nav>
  <div>&nbsp;</div>
  <div class="container text-center">
    <div class="row">
      <div class="col-xs-1 col-sm-2 col-md-3 col-lg-4">&nbsp;</div>
      <div class="form col-xs-10 col-sm-8 col-md-6 col-lg-4">
        {#if token == ""}
          {#if link == ""}
            <p>
              Input your secret here. It will be encrypted and saved to the
              server, and an one-time link will be generated.
            </p>
            <textarea
              class="form-control"
              id="secretPlace"
              style="height: 300px; font-family: monospace;"
              bind:value={contents}
            />
            <div>&nbsp;</div>
            <div class="input-group">
              <div class="input-group-prepend">
                <span class="input-group-text">Expires after</span>
              </div>
              <input
                type="number"
                class="form-control"
                aria-label="Default"
                aria-describedby="inputGroup-sizing-default"
                bind:value={expiryDays}
                min="1"
                max={initData.max_days}
              />
              <div class="input-group-append">
                <span class="input-group-text">days</span>
              </div>
            </div>
            <div>&nbsp;</div>
            <button
              type="button"
              class="btn btn-success"
              id="process"
              on:click={send}>Give me the link!</button
            >
          {:else}
            <label for="link" class="form-label"
              >Success! Your one-time link is:</label
            >
            <ClipboardableField id="link" text={link} />
            <hr />
            <label for="linkNoKey" class="form-label"
              >Or you can share the link without secret key:</label
            >
            <ClipboardableField id="linkNoKey" text={linkNoKey} />
            <br />
            <label for="linkNoKey" class="form-label"
              >And, separately, the key:</label
            >
            <ClipboardableField id="linkSecret" text={linkSecret} />
            <hr />
            <p>
              <i
                >Note: if the user inputs the wrong key, the secret will not be
                revealed but it will be "used" all the same, and the link will
                be invalid.</i
              >
            </p>
          {/if}
        {:else if contents == ""}
          <button
            type="button"
            class="btn btn-warning"
            id="peek"
            on:click={peek}>Is the secret still available?</button
          >
          <div>&nbsp;</div>
          <button
            type="button"
            class="btn btn-success"
            id="reveal"
            on:click={reveal}>Reveal the secret - One Time Only!</button
          >
        {:else}
          <label for="secretRevealed" class="form-label"
            >Success! Your secret is:</label
          >
          <textarea
            class="form-control"
            id="secretRevealed"
            style="height: 300px; font-family: monospace;"
            value={contents}
            readonly
            disabled
          />{/if}
      </div>
      <div class="col-xs-1 col-sm-2 col-md-3 col-lg-4">&nbsp;</div>
    </div>
  </div>
{/if}
