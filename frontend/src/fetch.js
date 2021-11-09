// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

export async function fetchJSON(url, opts = {}) {
  try {
    const res = await fetch(url, opts);
    const json = await res.json();
    if (res.ok) return [json, ""];
    return ["", json ? json : res.statusText];
  } catch (e) {
    return ["", e.message];
  }
}
