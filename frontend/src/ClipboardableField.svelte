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
    import { TOAST } from "./Utils.svelte";

    let { text, id } = $props();

    async function copy() {
        navigator.clipboard.writeText(text);
        await TOAST("Copied to clipboard.");
    }
</script>

<div class="input-group">
    <span
        class="input-group-text"
        id="basic-addon1"
        onclick={copy}
        onkeypress={copy}
        role="button"
        tabindex="0"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            fill="currentColor"
            class="bi bi-clipboard"
            viewBox="0 0 16 16"
        >
            <path
                d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"
            />
            <path
                d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"
            />
        </svg>
    </span>
    <input
        type="text"
        class="form-control"
        {id}
        value={text}
        readonly
        disabled
    />
</div>
