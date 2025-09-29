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

  let initData = $state(null);
  let token = $state("");
  let contents = $state("");
  let link = $state("");
  let linkNoKey = $state("");
  let linkSecret = $state("");
  let expiryDays = $state(3);
  let user = $state(null);
  let isLoggedIn = $state(false);
  let bypassLink = $state("");
  let bypassValidityHours = $state(1);
  let bypassToken = $state("");
  let selectedMode = $state("secret"); // "secret" or "bypass"

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

    const _bypassToken = getParameterByName("bt");
    bypassToken = !_bypassToken ? "" : _bypassToken;

    const ret = await CALL("getInitData", "GET");
    if (ret.isErr) {
      await ERROR(`Cannot load init data. ${ret.message}.`);
    } else {
      initData = ret.payload;
      expiryDays = initData.default_days;

      // Check user authentication if OAuth is enabled
      if (initData.oauth_enabled) {
        await checkAuthentication();
      }
    }

    // Initialize Bootstrap tooltips
    if (typeof window !== "undefined" && window.bootstrap) {
      const tooltipTriggerList = document.querySelectorAll(
        '[data-bs-toggle="tooltip"]',
      );
      tooltipTriggerList.forEach((tooltipTriggerEl) => {
        new window.bootstrap.Tooltip(tooltipTriggerEl);
      });
    }
  });

  async function checkAuthentication() {
    const ret = await CALL("auth/user", "GET");
    if (!ret.isErr && ret.payload.user) {
      user = ret.payload.user;
      isLoggedIn = true;

      // Reinitialize tooltips after login
      setTimeout(() => {
        if (typeof window !== "undefined" && window.bootstrap) {
          const tooltipTriggerList = document.querySelectorAll(
            '[data-bs-toggle="tooltip"]',
          );
          tooltipTriggerList.forEach((tooltipTriggerEl) => {
            new window.bootstrap.Tooltip(tooltipTriggerEl);
          });
        }
      }, 100);
    } else {
      user = null;
      isLoggedIn = false;
    }
  }

  function login() {
    window.location.href = "/api/auth/login";
  }

  async function logout() {
    const ret = await CALL("auth/logout", "POST");
    if (!ret.isErr) {
      user = null;
      isLoggedIn = false;
      await TOAST("Logged out successfully");
    } else {
      await ERROR(`Logout failed. ${ret.message}.`);
    }
  }

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

    // Include bypass token in query parameters if available
    const queryParams = bypassToken ? { bt: bypassToken } : null;
    const ret = await CALL("putSecret", "PUT", obj, queryParams);
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

  async function generateBypassLink() {
    if (typeof bypassValidityHours != "number")
      bypassValidityHours = parseInt(bypassValidityHours);

    if (
      bypassValidityHours < 1 ||
      bypassValidityHours > 24 ||
      isNaN(bypassValidityHours)
    ) {
      await ERROR("Invalid validity hours! Must be between 1 and 24.");
      bypassValidityHours = 1;
      return;
    }

    const obj = {
      validity_hours: bypassValidityHours,
    };
    const ret = await CALL("auth/generate-bypass-link", "POST", obj);
    if (ret.isErr) {
      await ERROR(`Bypass link generation failed. ${ret.message}.`);
    } else {
      bypassLink = ret.payload.bypass_url;
      await TOAST("Bypass link generated successfully!");
    }
  }

  function handleModeChange() {
    // Clear any existing generated links when switching modes
    bypassLink = "";
    link = "";
    linkNoKey = "";
    linkSecret = "";
  }

  function returnToMain() {
    // Only allow return to main in creation mode, not when viewing secrets
    if (token == "") {
      // Reset all state to return to main creation page
      contents = "";
      link = "";
      linkNoKey = "";
      linkSecret = "";
      bypassLink = "";
      selectedMode = "secret";
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
      <div class="d-flex align-items-center">
        {#if initData.oauth_enabled}
          {#if isLoggedIn && user}
            <!-- svelte-ignore a11y_invalid_attribute -->
            <a
              href="#"
              style="color: white;"
              onclick={logout}
              title="Logout"
              aria-label="Logout"
              data-bs-toggle="tooltip"
              data-bs-placement="bottom"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                fill="currentColor"
                viewBox="0 0 16 16"
              >
                <path
                  fill-rule="evenodd"
                  d="M10 12.5a.5.5 0 0 1-.5.5h-8a.5.5 0 0 1-.5-.5v-9a.5.5 0 0 1 .5-.5h8a.5.5 0 0 1 .5.5v2a.5.5 0 0 0 1 0v-2A1.5 1.5 0 0 0 9.5 2h-8A1.5 1.5 0 0 0 0 3.5v9A1.5 1.5 0 0 0 1.5 14h8a1.5 1.5 0 0 0 1.5-1.5v-2a.5.5 0 0 0-1 0v2z"
                />
                <path
                  fill-rule="evenodd"
                  d="M15.854 8.354a.5.5 0 0 0 0-.708l-3-3a.5.5 0 0 0-.708.708L14.293 7.5H5.5a.5.5 0 0 0 0 1h8.793l-2.147 2.146a.5.5 0 0 0 .708.708l3-3z"
                />
              </svg>
            </a>&nbsp;&nbsp;
          {/if}
        {/if}
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
          {#if link == "" && bypassLink == ""}
            {#if initData.oauth_enabled && !isLoggedIn && !bypassToken}
              <div class="text-center py-5">
                <div class="mb-4">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="64"
                    height="64"
                    fill="#000000"
                    viewBox="0 0 256 256"
                    ><path
                      d="M128,20A108,108,0,1,0,236,128,108.12,108.12,0,0,0,128,20Zm0,192a84,84,0,1,1,84-84A84.09,84.09,0,0,1,128,212Zm0-144a44,44,0,0,0-33.61,72.41l-9.86,32.06A12,12,0,0,0,96,188h64a12,12,0,0,0,11.47-15.53l-9.86-32.06A44,44,0,0,0,128,68Zm8.53,72.51L143.75,164h-31.5l7.22-23.49a12,12,0,0,0-4-12.89,20,20,0,1,1,25,0A12,12,0,0,0,136.53,140.51Z"
                    ></path></svg
                  >
                </div>
                <h5 class="text-muted mb-3">Authentication Required</h5>
                <p class="lead text-muted">
                  Please
                  <button
                    type="button"
                    onclick={login}
                    class="btn btn-link p-0 fw-bold text-decoration-none"
                    style="color: #667eea; vertical-align: baseline;"
                  >
                    login
                  </button>
                  to create secrets.<br /><br />The secrets will be accessible
                  one time, only by the recipient of the link.
                </p>
              </div>
            {:else}
              {#if bypassToken && initData.oauth_enabled && !isLoggedIn}
                <div class="alert alert-info" role="alert">
                  <small>
                    <strong>✨ Guest Access:</strong> You're creating a secret using
                    a one-time invitation link. This access is temporary and will
                    expire soon.
                  </small>
                </div>
              {/if}

              <!-- Mode Selection Radio Buttons (only show if both modes are available) -->
              {#if initData.oauth_enabled && initData.allow_bypass_link && isLoggedIn && !bypassToken}
                <div class="text-center mb-4">
                  <div
                    class="btn-group"
                    role="group"
                    aria-label="Mode selection"
                  >
                    <input
                      type="radio"
                      class="btn-check"
                      name="mode"
                      id="secretMode"
                      bind:group={selectedMode}
                      value="secret"
                      onchange={handleModeChange}
                    />
                    <label
                      class="btn btn-outline-success btn-small"
                      for="secretMode"
                    >
                      Create Secret Link
                    </label>

                    <input
                      type="radio"
                      class="btn-check"
                      name="mode"
                      id="bypassMode"
                      bind:group={selectedMode}
                      value="bypass"
                      onchange={handleModeChange}
                    />
                    <label
                      class="btn btn-outline-success btn-small"
                      for="bypassMode"
                    >
                      Generate Access Link
                    </label>
                  </div>
                </div>
              {/if}

              <!-- Secret Creation Mode -->
              {#if selectedMode === "secret" || bypassToken || !initData.oauth_enabled || !initData.allow_bypass_link || !isLoggedIn}
                <p class="small text-muted mb-4">
                Input your secret here. It will be encrypted and saved to the
                server, and an one-time link will be generated.
              </p>
              <textarea
                class="form-control"
                id="secretPlace"
                style="height: 300px; font-family: monospace;"
                bind:value={contents}
              ></textarea>
              <div>&nbsp;</div>
              <div class="input-group">
                <div class="input-group-prepend">
                    <span class="input-group-text text-muted"
                      >Expires after</span
                    >
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
                    <span class="input-group-text text-muted">days</span>
                </div>
              </div>
              <div>&nbsp;</div>
              <button
                type="button"
                class="btn btn-success"
                id="process"
                onclick={send}>Give me the link!</button
              >
            {/if}

              <!-- Bypass Link Generation Mode -->
              {#if selectedMode === "bypass" && initData.oauth_enabled && initData.allow_bypass_link && isLoggedIn}
                <div class="text-center">
                  <p class="small text-muted mb-4">
                    Create a one-time link that allows unauthenticated users to
                    create secrets.
                  </p>

                  <div class="input-group mb-3">
                    <div class="input-group-prepend">
                      <span class="input-group-text text-muted">Valid for</span>
                    </div>
                    <input
                      type="number"
                      class="form-control"
                      aria-label="Validity Hours"
                      bind:value={bypassValidityHours}
                      min="1"
                      max="24"
                    />
                    <div class="input-group-append">
                      <span class="input-group-text text-muted">hours</span>
                    </div>
                  </div>

                  <button
                    type="button"
                    class="btn btn-success"
                    onclick={generateBypassLink}
                  >
                    Generate Access Link
                  </button>
                </div>
              {/if}
            {/if}
          {:else if link != ""}
            <!-- Secret Link Results -->
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
          {:else if bypassLink != ""}
            <!-- Bypass Link Results -->
            <label for="bypassLink" class="form-label">
              <strong
                >Access Link Generated<br />(single-use, expires in {bypassValidityHours}
                hour{bypassValidityHours === 1 ? "" : "s"})</strong
              >
            </label>
            <ClipboardableField id="bypassLink" text={bypassLink} />
            <p class="small text-muted mt-2">
              <i
                >Share this link with unauthenticated users to allow them to
                create secrets without logging in.</i
              >
            </p>
          {/if}
        {:else if contents == ""}
          <button type="button" class="btn btn-warning" id="peek" onclick={peek}
            >Is the secret still available?</button
          >
          <div>&nbsp;</div>
          <button
            type="button"
            class="btn btn-success"
            id="reveal"
            onclick={reveal}>Reveal the secret - One Time Only!</button
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
          ></textarea>{/if}
      </div>
      <div class="col-xs-1 col-sm-2 col-md-3 col-lg-4">&nbsp;</div>
    </div>
  </div>
{/if}
