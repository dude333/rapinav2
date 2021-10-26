// SPDX - FileCopyrightText: 2021 Adriano Prado < dev@dude333.com>
//
// SPDX - License - Identifier: MIT;

import { fetchJSON } from "./fetch";

export async function apiDFP(cnpj, ano) {
  const url = `/api/dfp?cnpj=${cnpj}`;
  return fetchJSON(url);
}
