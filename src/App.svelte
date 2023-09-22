<script>
  import ClipboardableField from "./ClipboardableField.svelte";
  import { CALL, DECRYPT, ENCRYPT, ERROR, TOAST } from "./Utils.svelte";
  import { onMount } from "svelte";

  $: token = "";
  $: contents = "";
  $: link = "";
  $: linkNoKey = "";
  $: linkSecret = "";
  $: expiryDays = 7;

  function getParameterByName(name, url = window.location.href) {
    name = name.replace(/[\[\]]/g, "\\$&");
    var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
      results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return "";
    return decodeURIComponent(results[2].replace(/\+/g, " "));
  }

  onMount(() => {
    const _token = getParameterByName("t");
    token = !_token ? "" : _token;
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

    const encoded = ENCRYPT(contents);

    const req = {
      transaction: [
        {
          query: "^S1",
          values: {
            iv: encoded.iv,
            secret: encoded.text,
            sha: encoded.sha,
            expiry: expiryDays,
          },
        },
      ],
    };

    const ret = await CALL(req);
    if (ret.status == 200) {
      linkNoKey =
        location.protocol +
        "//" +
        location.host +
        "?t=" +
        ret.results[0].resultSet[0]["ID"];
      link = linkNoKey + "&s=" + encodeURIComponent(encoded.key);
      linkSecret = encoded.key;
    } else {
      await ERROR(JSON.stringify(ret));
    }
  }

  async function peek() {
    const req = {
      transaction: [
        {
          query: "^S2",
          values: { id: token },
        },
      ],
    };

    const ret = await CALL(req);
    if (ret.status != 200) await ERROR(JSON.stringify(ret));
    else if (ret.results[0].resultSet.length == 0)
      await TOAST("Secret expired, already revealed or wrong link.");
    else await TOAST("Secret (still) available.");
  }

  async function reveal() {
    let key = getParameterByName("s");
    if (!key) {
      key = prompt("Decoding key").trim();
    }

    const req = {
      transaction: [
        {
          query: "^S3",
          values: { id: token },
        },
      ],
    };

    const ret = await CALL(req);
    if (ret.status != 200) {
      await ERROR(JSON.stringify(ret));
      return;
    }

    if (ret.results[0].resultSet.length == 0) {
      await ERROR("Secret expired, already revealed or wrong link.");
      return;
    }

    const encoded = {
      key,
      iv: ret.results[0].resultSet[0]["IV"],
      text: ret.results[0].resultSet[0]["SECRET"],
      sha: ret.results[0].resultSet[0]["SHA"],
    };
    try {
      contents = DECRYPT(encoded);
    } catch (e) {
      await ERROR(e);
    }
  }
</script>

<nav class="navbar navbar-expand-lg bg-success text-white">
  <div class="container-fluid bg-success text-white">
    <div class="navbar-brand bg-success text-white">
      🔐 Seif
      <span class="small"
        ><small><small>&nbsp;one time secrets storage - v0.0.1</small></small
        ></span
      >
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
              revealed but it will be "used" all the same, and the link will be
              invalid.</i
            >
          </p>
        {/if}
      {:else if contents == ""}
        <button type="button" class="btn btn-warning" id="peek" on:click={peek}
          >Is the secret still available?</button
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

<style>
</style>
