"use strict";

const state = { cases: [], view: null, currentCaseID: null };
const statusNames = { draft:"草稿建档", baseline_frozen:"基线已冻结", deliberation:"公开评议", pending_review:"待独立复核", changes_requested:"定向退回", approved:"复核通过", sealed:"已定稿" };
const evidenceNames = { edge_match:"边缘对应", fabric_match:"胎土特征", decoration_continuity:"纹饰连续性", scale_measurements:"尺度测量", image_refs:"图像引用" };
const eventNames = { "case.created":"创建案件", "sherd.added":"登记陶片", "baseline.frozen":"冻结基线", "hypothesis.added":"登记候选", "hypothesis.published":"公开候选", "challenge.raised":"提出异议", "challenge.resolved":"处置异议", "case.review_requested":"提交复核", "case.reviewed":"完成复核", "case.sealed":"定稿封存" };

const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => Array.from(document.querySelectorAll(selector));
const escapeHTML = (value) => String(value ?? "").replace(/[&<>'"]/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"})[char]);
const requestID = () => `web-${Date.now()}-${crypto.getRandomValues(new Uint32Array(1))[0].toString(16)}`;

async function api(path, options = {}) {
  const response = await fetch(path, { headers:{"Content-Type":"application/json", ...(options.headers || {})}, ...options });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.error?.message || `请求失败（${response.status}）`);
  return payload;
}

function meta(actorID) { return { request_id:requestID(), expected_revision:state.view.case.revision, actor_id:actorID }; }
function formObject(form) { return Object.fromEntries(new FormData(form).entries()); }

function showNotice(message, error = false) {
  const node = $("#notice"); node.textContent = message; node.classList.remove("hidden", "error");
  if (error) node.classList.add("error");
  window.setTimeout(() => node.classList.add("hidden"), 5000);
}

async function refreshCases() {
  const payload = await api("/api/cases"); state.cases = payload.cases; renderCaseList();
}

function renderCaseList() {
  $("#case-list").innerHTML = state.cases.length ? state.cases.map(item => `<button class="case-item ${item.case_id === state.currentCaseID ? "active" : ""}" data-case-id="${escapeHTML(item.case_id)}"><strong>${escapeHTML(item.site_unit)} · ${escapeHTML(item.vessel_class)}</strong><span>${escapeHTML(item.case_id)} · ${escapeHTML(statusNames[item.status] || item.status)}</span></button>`).join("") : `<div class="inline-empty">暂无案件</div>`;
  $$(".case-item").forEach(button => button.addEventListener("click", () => openCase(button.dataset.caseId)));
}

async function openCase(caseID) {
  try {
    state.view = await api(`/api/cases/${encodeURIComponent(caseID)}`); state.currentCaseID = caseID;
    $("#empty-state").classList.add("hidden"); $("#case-workspace").classList.remove("hidden");
    renderCaseList(); renderView();
  } catch (error) { showNotice(error.message, true); }
}

function renderView() {
  const { case:c, sherds, hypotheses, challenges, actions, timeline } = state.view;
  $("#case-id").textContent = c.case_id; $("#case-title").textContent = `${c.site_unit} · ${c.vessel_class}`;
  $("#case-meta").textContent = `负责人 ${c.owner_id} · 创建于 ${new Date(c.created_at).toLocaleString("zh-CN")}`;
  $("#status-badge").textContent = statusNames[c.status] || c.status; $("#revision").textContent = `revision ${c.revision}`;
  renderStages(c.status); renderSherds(sherds); renderHypotheses(hypotheses); renderChallenges(challenges, hypotheses); renderReviews(c.reviews || []); renderTimeline(timeline);
  $("#sherd-form").classList.toggle("hidden", !actions.can_add_sherd); $("#freeze-button").classList.toggle("hidden", !actions.can_freeze_baseline);
  $("#hypothesis-form").classList.toggle("hidden", !actions.can_add_hypothesis); $("#challenge-form").classList.toggle("hidden", !actions.can_challenge);
  $("#returned-evidence-form").classList.toggle("hidden", !actions.can_revise_returned); renderReturnedEvidenceForm(c, hypotheses);
  $("#request-review-button").classList.toggle("hidden", !actions.can_request_review); $("#review-form").classList.toggle("hidden", !actions.can_review);
  $("#finalize-form").classList.toggle("hidden", !actions.can_finalize); $("#verify-button").classList.toggle("hidden", c.status !== "sealed");
  if (c.status === "sealed") loadDossier(); else { $("#dossier-view").classList.add("hidden"); $("#dossier-empty").classList.remove("hidden"); }
}

function renderStages(status) {
  const order = ["draft","baseline_frozen","deliberation","pending_review","sealed"];
  const mapped = status === "changes_requested" ? "baseline_frozen" : status === "approved" ? "pending_review" : status;
  const index = order.indexOf(mapped);
  $$(".status-track span").forEach((node, position) => { node.classList.toggle("done", position < index); node.classList.toggle("current", position === index); });
}

function renderSherds(sherds) {
  $("#sherd-rows").innerHTML = sherds.length ? sherds.map(s => `<tr><td><strong>${escapeHTML(s.sherd_id)}</strong></td><td>${escapeHTML(s.context_code)}</td><td>${escapeHTML(s.fabric_code)}</td><td>${escapeHTML(s.rim_profile)}</td><td>${s.dimensions_mm.height} × ${s.dimensions_mm.width} × ${s.dimensions_mm.depth}</td><td>${escapeHTML(s.image_ref)}<br><code>${escapeHTML(s.image_digest.slice(0,12))}…</code></td></tr>`).join("") : `<tr><td colspan="6" class="muted">尚未登记陶片</td></tr>`;
}

function evidenceText(evidence, key) {
  const value = evidence[key];
  if (key === "scale_measurements") return Object.entries(value || {}).map(([name, number]) => `${name}: ${number}`).join("；");
  if (key === "image_refs") return (value || []).join("；");
  return value || "缺失";
}

function renderHypotheses(hypotheses) {
  $("#hypothesis-list").innerHTML = hypotheses.length ? hypotheses.map(h => `<article class="matrix"><div class="matrix-head"><div><h3>${escapeHTML(h.hypothesis_id)} <span class="tag">${escapeHTML(h.status)}</span></h3><span class="muted">${h.sherd_ids.map(escapeHTML).join(" + ")} · 作者 ${escapeHTML(h.author_id)} · 证据 v${h.evidence_version}</span></div><div><span class="score">完整度 ${h.completeness}%</span>${h.status === "draft" ? `<button class="primary submit-hypothesis" data-hypothesis-id="${escapeHTML(h.hypothesis_id)}" data-author-id="${escapeHTML(h.author_id)}">公开评议</button>` : ""}</div></div><div class="matrix-grid">${Object.keys(evidenceNames).map(key => `<div class="evidence-cell ${h.locked_keys?.[key] ? "locked" : ""}"><strong>${evidenceNames[key]} ${h.locked_keys?.[key] ? "· 已锁定" : ""}</strong><span>${escapeHTML(evidenceText(h.evidence,key))}</span></div>`).join("")}</div></article>`).join("") : `<div class="inline-empty">尚未登记候选拼合</div>`;
  $$(".submit-hypothesis").forEach(button => button.addEventListener("click", () => submitHypothesis(button.dataset.hypothesisId, button.dataset.authorId)));
}

function renderReturnedEvidenceForm(caseData, hypotheses) {
  const form = $("#returned-evidence-form");
  form.querySelector('[name="hypothesis_id"]').innerHTML = hypotheses.filter(item => item.status !== "withdrawn").map(item => `<option value="${escapeHTML(item.hypothesis_id)}">${escapeHTML(item.hypothesis_id)}</option>`).join("");
  form.querySelector('[name="evidence_key"]').innerHTML = Object.keys(caseData.reopened_keys || {}).filter(key => caseData.reopened_keys[key]).map(key => `<option value="${escapeHTML(key)}">${escapeHTML(evidenceNames[key] || key)}</option>`).join("");
}

function renderChallenges(challenges, hypotheses) {
  const published = hypotheses.filter(item => item.status === "published");
  $("#challenge-form select[name=hypothesis_id]").innerHTML = published.map(item => `<option value="${escapeHTML(item.hypothesis_id)}">${escapeHTML(item.hypothesis_id)}</option>`).join("");
  $("#challenge-list").innerHTML = challenges.length ? challenges.map(ch => `<article class="queue-item"><div><h3>${escapeHTML(ch.challenge_id)} <span class="tag">${ch.status === "open" ? "待处置" : "已关闭"}</span></h3><div class="queue-meta">${escapeHTML(ch.hypothesis_id)} · ${escapeHTML(evidenceNames[ch.evidence_key])} · ${escapeHTML(ch.raised_by)}</div><p>${escapeHTML(ch.statement)}</p>${ch.status === "closed" ? `<p><strong>${escapeHTML(ch.resolution_kind)}</strong>：${escapeHTML(ch.resolution_note)}</p>` : ""}</div>${ch.status === "open" ? `<form class="resolution-form" data-challenge-id="${escapeHTML(ch.challenge_id)}" data-evidence-key="${escapeHTML(ch.evidence_key)}"><label>处置动作<select name="kind"><option value="maintain">维持答复</option><option value="supplement">补充证据</option><option value="withdraw">撤回候选</option></select></label><label>处置人<input name="actor_id" required></label><label>处置说明<input name="note" required></label><label>补充内容<input name="replacement" placeholder="补证时填写"></label><button class="secondary" type="submit">完成处置</button></form>` : ""}</article>`).join("") : `<div class="inline-empty">暂无异议</div>`;
  $$(".resolution-form").forEach(form => form.addEventListener("submit", resolveChallenge));
}

function renderReviews(reviews) {
  $("#review-history").innerHTML = reviews.length ? reviews.map(review => `<div class="review-card"><strong>${review.decision === "approve" ? "通过" : "定向退回"} · ${escapeHTML(review.reviewer_id)}</strong><p>${escapeHTML(review.reason)}</p><span class="queue-meta">${new Date(review.created_at).toLocaleString("zh-CN")}${review.reopen_keys?.length ? ` · 开放 ${review.reopen_keys.map(key => evidenceNames[key] || key).join("、")}` : ""}</span></div>`).join("") : `<div class="inline-empty">尚无复核记录</div>`;
}

function renderTimeline(timeline) {
  $("#timeline-list").innerHTML = timeline.map(item => `<li><strong>${escapeHTML(eventNames[item.kind] || item.kind)}</strong><span>${escapeHTML(item.actor_id)} · ${escapeHTML(item.occurred_at)}</span><code>${escapeHTML(item.digest)}</code></li>`).join("");
}

async function mutate(path, body, success) {
  try { state.view = await api(path, {method:"POST", body:JSON.stringify(body)}); renderView(); await refreshCases(); showNotice(success); }
  catch (error) { showNotice(error.message, true); }
}

async function submitHypothesis(id, actor) { await mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/hypotheses/${encodeURIComponent(id)}/submit`, meta(actor), "候选已进入公开评议"); }

async function resolveChallenge(event) {
  event.preventDefault(); const form = event.currentTarget; const values = formObject(form); const kind = values.kind; const key = form.dataset.evidenceKey;
  let replacement;
  if (kind === "supplement") {
    if (key === "scale_measurements") replacement = parseMeasurements(values.replacement);
    else if (key === "image_refs") replacement = values.replacement.split(",").map(item => item.trim()).filter(Boolean);
    else replacement = values.replacement;
  }
  await mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/challenges/${encodeURIComponent(form.dataset.challengeId)}/resolve`, {...meta(values.actor_id), resolution_kind:kind, resolution_note:values.note, ...(kind === "supplement" ? {replacement} : {})}, "异议已完成处置");
}

function parseMeasurements(text) {
  const result = {}; String(text).split(",").forEach(item => { const [key,value] = item.split(":"); if (key?.trim()) result[key.trim()] = Number(value); }); return result;
}

async function loadDossier() {
  try { renderDossier(await api(`/api/cases/${encodeURIComponent(state.currentCaseID)}/dossier`)); }
  catch (error) { showNotice(error.message, true); }
}

function decodePayload(base64) {
  try { const bytes = Uint8Array.from(atob(base64), char => char.charCodeAt(0)); return JSON.stringify(JSON.parse(new TextDecoder().decode(bytes)), null, 2); } catch (_) { return base64; }
}

function renderDossier(view) {
  const dossier = view.dossier; $("#dossier-empty").classList.add("hidden"); $("#dossier-view").classList.remove("hidden");
  const verify = $("#verify-result"); verify.textContent = `${view.valid ? "校验通过" : "校验失败"}：${view.message}`; verify.classList.toggle("invalid", !view.valid);
  $("#dossier-id").textContent = dossier.dossier_id; $("#approved-hypotheses").textContent = dossier.approved_hypothesis_ids.join("、"); $("#dossier-reviewer").textContent = `${dossier.reviewer_id} · ${dossier.review_reason}`;
  $("#chain-head").textContent = dossier.event_chain_head; $("#dossier-sha").textContent = dossier.sha256; $("#canonical-payload").textContent = decodePayload(dossier.canonical_payload);
}

function activateTab(name) { $$(".tab").forEach(tab => tab.classList.toggle("active", tab.dataset.tab === name)); $$(".tab-panel").forEach(panel => panel.classList.toggle("hidden", panel.id !== `tab-${name}`)); }
function openNewCase() { $("#case-dialog").showModal(); }

$("#new-case-button").addEventListener("click", openNewCase); $$('[data-action="new-case"]').forEach(button => button.addEventListener("click", openNewCase));
$$(".tab").forEach(tab => tab.addEventListener("click", () => activateTab(tab.dataset.tab)));

$("#case-form").addEventListener("submit", async event => {
  event.preventDefault(); const values = formObject(event.currentTarget);
  try { const view = await api("/api/cases", {method:"POST",body:JSON.stringify({...values,request_id:requestID(),expected_revision:0})}); $("#case-dialog").close(); event.currentTarget.reset(); await refreshCases(); state.currentCaseID = view.case.case_id; state.view = view; $("#empty-state").classList.add("hidden"); $("#case-workspace").classList.remove("hidden"); renderView(); renderCaseList(); }
  catch (error) { $("#case-dialog").close(); showNotice(error.message,true); }
});

$("#sherd-form").addEventListener("submit", async event => { event.preventDefault(); const v=formObject(event.currentTarget); const sherd={sherd_id:v.sherd_id,context_code:v.context_code,fabric_code:v.fabric_code,rim_profile:v.rim_profile,dimensions_mm:{height:Number(v.height),width:Number(v.width),depth:Number(v.depth)},image_ref:v.image_ref,image_digest:v.image_digest}; await mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/sherds`,{...meta(state.view.case.owner_id),sherd},"陶片已登记"); event.currentTarget.reset(); });
$("#freeze-button").addEventListener("click",()=>mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/freeze`,meta(state.view.case.owner_id),"陶片基线已冻结"));
$("#hypothesis-form").addEventListener("submit",async event=>{event.preventDefault();const v=formObject(event.currentTarget);const evidence={edge_match:v.edge_match,fabric_match:v.fabric_match,decoration_continuity:v.decoration_continuity,scale_measurements:parseMeasurements(v.scale_measurements),image_refs:v.image_refs.split(",").map(x=>x.trim()).filter(Boolean)};await mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/hypotheses`,{...meta(v.actor_id),hypothesis_id:v.hypothesis_id,sherd_ids:v.sherd_ids.split(",").map(x=>x.trim()).filter(Boolean),evidence},"候选拼合已保存");event.currentTarget.reset();});
$("#returned-evidence-form").addEventListener("submit",async event=>{event.preventDefault();const v=formObject(event.currentTarget);let replacement=v.replacement;if(v.evidence_key==="scale_measurements")replacement=parseMeasurements(v.replacement);if(v.evidence_key==="image_refs")replacement=v.replacement.split(",").map(x=>x.trim()).filter(Boolean);await mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/hypotheses/${encodeURIComponent(v.hypothesis_id)}/evidence/${encodeURIComponent(v.evidence_key)}/revise`,{...meta(v.actor_id),note:v.note,replacement},"退回证据已生成新版本，请重新公开评议");event.currentTarget.reset();});
$("#challenge-form").addEventListener("submit",async event=>{event.preventDefault();const v=formObject(event.currentTarget);await mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/challenges`,{...meta(v.actor_id),challenge_id:v.challenge_id,hypothesis_id:v.hypothesis_id,evidence_key:v.evidence_key,statement:v.statement},"异议已登记");event.currentTarget.reset();});
$("#request-review-button").addEventListener("click",()=>mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/request-review`,meta(state.view.case.owner_id),"案件已提交独立复核"));
$("#review-form").addEventListener("submit",async event=>{event.preventDefault();const v=formObject(event.currentTarget);await mutate(`/api/cases/${encodeURIComponent(state.currentCaseID)}/review`,{request_id:requestID(),expected_revision:state.view.case.revision,reviewer_id:v.reviewer_id,decision:v.decision,reason:v.reason,reopen_keys:v.reopen_keys.split(",").map(x=>x.trim()).filter(Boolean)},"复核结论已提交");event.currentTarget.reset();});
$("#finalize-form").addEventListener("submit",async event=>{event.preventDefault();const v=formObject(event.currentTarget);try{const result=await api(`/api/cases/${encodeURIComponent(state.currentCaseID)}/finalize`,{method:"POST",body:JSON.stringify({...meta(v.actor_id),dossier_id:v.dossier_id})});renderDossier(result);await openCase(state.currentCaseID);activateTab("dossier");showNotice("案件已冻结并生成只读档案");}catch(error){showNotice(error.message,true);}});
$("#verify-button").addEventListener("click",async()=>{try{renderDossier(await api(`/api/cases/${encodeURIComponent(state.currentCaseID)}/dossier/verify`));}catch(error){showNotice(error.message,true);}});

(async function start(){try{await api("/healthz");$("#connection").textContent="存储已连接";$("#connection").classList.add("ok");await refreshCases();}catch(error){$("#connection").textContent="服务不可用";showNotice(error.message,true);}})();
