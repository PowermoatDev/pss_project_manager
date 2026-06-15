<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-mark">P</div>
        <div>
          <div class="brand-title">列印資安專案戰情室</div>
          <div class="brand-subtitle">Vue + Go + MSSQL</div>
        </div>
      </div>
      <button class="nav active">總覽儀表板</button>
      <button class="nav">專案管理</button>
      <button class="nav">報價管理</button>
      <button class="nav">客製需求</button>
      <button class="nav">共用行事曆</button>
      <div class="sidebar-foot">
        API：Go<br />
        DB：Microsoft SQL Server<br />
        檔案：uploads/
      </div>
    </aside>

    <main class="main">
      <header class="topbar">
        <label class="search">
          <span>搜尋</span>
          <input v-model.trim="query" placeholder="搜尋專案、規格、報價、POC 結果..." />
        </label>
        <div class="status" :class="{ error: hasError }">{{ syncText }}</div>
        <button class="primary" @click="openProjectModal">新增專案</button>
      </header>

      <section class="layout">
        <aside class="panel stage-panel">
          <div class="section-head">
            <h2>流程狀態</h2>
            <button class="ghost" @click="loadAll">同步</button>
          </div>
          <button v-for="stage in activeStages" :key="stage.key" class="stage-card" :style="{ '--accent': stage.color }" @click="setView('active', stage.key)">
            <span>{{ stage.label }}</span>
            <strong>{{ countStage(stage.key) }}</strong>
            <small>進行中專案</small>
          </button>
          <div class="divider">管理區域</div>
          <button class="stage-card" style="--accent:#52ca78" @click="setView('done', 'all')">
            <span>已完成結案瀏覽</span>
            <strong>{{ doneCount }}</strong>
            <small>結案且有預計安裝日期</small>
          </button>
          <button class="stage-card" style="--accent:#a7b0b8" @click="setView('lost', 'all')">
            <span>未成案瀏覽</span>
            <strong>{{ lostCount }}</strong>
            <small>結案但未填預計安裝日期</small>
          </button>
        </aside>

        <section class="workspace">
          <section class="panel metrics">
            <div>
              <small>進行中</small>
              <strong>{{ activeCount }}</strong>
            </div>
            <div>
              <small>本週行程</small>
              <strong>{{ weekEvents.length }}</strong>
            </div>
            <div>
              <small>待結案</small>
              <strong>{{ countStage('closing') }}</strong>
            </div>
            <div>
              <small>最後同步</small>
              <strong>{{ lastSync }}</strong>
            </div>
          </section>

          <section class="panel table-panel">
            <div class="section-head">
              <h2>{{ viewTitle }}</h2>
              <div class="toolbar">
                <button v-for="item in tabStages" :key="item.key" class="tab" :class="{ active: stageFilter === item.key && view === 'active' }" @click="setView('active', item.key)">
                  {{ item.label }}
                </button>
              </div>
            </div>
            <div class="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>專案名稱</th>
                    <th>規格內容</th>
                    <th>報價內容</th>
                    <th>報價日期 / 說明 / 檔案</th>
                    <th>客製化需求</th>
                    <th>客製化日期 / 說明 / 檔案</th>
                    <th>人天</th>
                    <th>POC 日期</th>
                    <th>POC 結果</th>
                    <th>預計安裝日期</th>
                    <th>完成日</th>
                    <th>備注說明</th>
                    <th>結案</th>
                    <th>狀態</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="project in visibleProjects" :key="project.id">
                    <td>
                      <input v-model="project.name" @change="saveProject(project)" />
                      <small>{{ project.code }}</small>
                    </td>
                    <td><textarea v-model="project.spec" @change="saveProject(project)" /></td>
                    <td><textarea v-model="project.quoteContent" @change="saveProject(project)" /></td>
                    <td>
                      <input type="date" v-model="project.quoteDate" @change="saveProject(project)" />
                      <textarea v-model="project.quoteNote" @change="saveProject(project)" />
                      <FileLine :file-name="project.quoteFileName" :file-url="project.quoteFileUrl" />
                      <input type="file" @change="uploadFile(project, 'quote', $event)" />
                    </td>
                    <td><textarea v-model="project.customNeed" @change="saveProject(project)" /></td>
                    <td>
                      <input type="date" v-model="project.customDate" @change="saveProject(project)" />
                      <textarea v-model="project.customNote" @change="saveProject(project)" />
                      <FileLine :file-name="project.customFileName" :file-url="project.customFileUrl" />
                      <input type="file" @change="uploadFile(project, 'custom', $event)" />
                    </td>
                    <td><input type="number" min="0" v-model.number="project.customDays" @change="saveProject(project)" /></td>
                    <td><input type="date" v-model="project.pocDate" @change="saveProject(project)" /></td>
                    <td><textarea v-model="project.pocResult" @change="saveProject(project)" /></td>
                    <td><input type="date" v-model="project.installDate" @change="saveProject(project)" /></td>
                    <td><input type="date" v-model="project.doneDate" @change="saveProject(project)" /></td>
                    <td><textarea v-model="project.latestNote" @change="saveProject(project)" /></td>
                    <td><input type="checkbox" :checked="project.isClosed" @change="closeProject(project)" /></td>
                    <td><span class="pill" :class="project.status">{{ labelStatus(project.status) }}</span></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </section>

        <aside class="panel calendar">
          <div class="section-head">
            <h2>共用行事曆</h2>
            <button class="primary" @click="openEventModal()">新增活動</button>
          </div>
          <div class="week-title">{{ weekTitle }}</div>
          <div class="calendar-grid">
            <div class="time-head"></div>
            <div v-for="day in weekDays" :key="day.key" class="day-head">{{ day.label }}</div>
            <template v-for="time in calendarTimes" :key="time">
              <div class="time-cell">{{ time }}</div>
              <div v-for="day in weekDays" :key="day.key + time" class="slot">
                <button v-for="event in eventsAt(day.key, time)" :key="event.id" class="event" :style="{ borderColor: event.color, background: event.color + '30' }" @click="openEventModal(event)">
                  <strong>{{ event.eventTime }} {{ event.type }}</strong>
                  <span>{{ event.title }}</span>
                  <small>{{ event.owner }} 編輯</small>
                </button>
              </div>
            </template>
          </div>
          <div class="upcoming">
            <h3>即將到來</h3>
            <button v-for="event in upcomingEvents" :key="event.id" class="upcoming-row" @click="openEventModal(event)">
              <span>{{ event.eventDate }} {{ event.eventTime }}</span>
              <strong>{{ event.type }}</strong>
              <span>{{ event.title }}</span>
            </button>
          </div>
        </aside>
      </section>
    </main>

    <dialog ref="projectDialog" class="dialog">
      <form method="dialog" class="dialog-body" @submit.prevent="createProject">
        <div class="section-head">
          <h2>新增列印資安專案</h2>
          <button class="ghost" type="button" @click="projectDialog?.close()">關閉</button>
        </div>
        <div class="form-grid">
          <label>專案名稱<input v-model="projectDraft.name" required /></label>
          <label>規格內容<input v-model="projectDraft.spec" /></label>
          <label>客戶名稱<input v-model="projectDraft.client" /></label>
          <label>報價內容<input v-model="projectDraft.quoteContent" /></label>
          <label>報價日期<input type="date" v-model="projectDraft.quoteDate" /></label>
          <label>報價說明<input v-model="projectDraft.quoteNote" /></label>
          <label>客製化日期<input type="date" v-model="projectDraft.customDate" /></label>
          <label>客製化人天<input type="number" min="0" v-model.number="projectDraft.customDays" /></label>
          <label>預計安裝日期<input type="date" v-model="projectDraft.installDate" /></label>
          <label class="wide">客製化需求<textarea v-model="projectDraft.customNeed" /></label>
          <label class="wide">POC 結果說明<textarea v-model="projectDraft.pocResult" /></label>
          <label class="wide">備注說明<textarea v-model="projectDraft.latestNote" /></label>
        </div>
        <div class="dialog-actions">
          <button class="ghost" type="button" @click="projectDialog?.close()">取消</button>
          <button class="primary" type="submit">建立專案</button>
        </div>
      </form>
    </dialog>

    <dialog ref="eventDialog" class="dialog">
      <form method="dialog" class="dialog-body" @submit.prevent="saveEvent">
        <div class="section-head">
          <h2>{{ eventDraft.id ? '編輯行事曆活動' : '新增行事曆活動' }}</h2>
          <button class="ghost" type="button" @click="eventDialog?.close()">關閉</button>
        </div>
        <div class="form-grid">
          <label>日期<input type="date" v-model="eventDraft.eventDate" required /></label>
          <label>時間<input type="time" v-model="eventDraft.eventTime" required /></label>
          <label>類型<input v-model="eventDraft.type" required /></label>
          <label class="wide">活動內容<input v-model="eventDraft.title" required /></label>
          <label>負責同仁<input v-model="eventDraft.owner" required /></label>
          <label>顏色<input type="color" v-model="eventDraft.color" /></label>
          <label>編輯者<input v-model="eventDraft.editor" /></label>
        </div>
        <div class="dialog-actions">
          <button v-if="eventDraft.id" class="danger" type="button" @click="deleteEvent">刪除</button>
          <button class="ghost" type="button" @click="eventDialog?.close()">取消</button>
          <button class="primary" type="submit">儲存</button>
        </div>
      </form>
    </dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import FileLine from "./components/FileLine.vue";

const projectDialog = ref(null);
const eventDialog = ref(null);
const projects = ref([]);
const events = ref([]);
const query = ref("");
const view = ref("active");
const stageFilter = ref("all");
const syncText = ref("連線中...");
const hasError = ref(false);
const lastSync = ref("-");

const stages = {
  need: { label: "需求確認", color: "#25d7ef" },
  quote: { label: "報價中", color: "#f7aa35" },
  poc: { label: "POC排程", color: "#5a8cff" },
  dev: { label: "客製開發", color: "#9a73ff" },
  closing: { label: "待結案", color: "#ff635f" },
  done: { label: "已完成結案", color: "#52ca78" },
  lost: { label: "未成案", color: "#a7b0b8" }
};
const activeStages = Object.entries(stages).filter(([key]) => !["done", "lost"].includes(key)).map(([key, value]) => ({ key, ...value }));
const tabStages = [{ key: "all", label: "全部" }, ...activeStages];
const today = () => new Date().toISOString().slice(0, 10);
const addDays = (date, days) => {
  const next = new Date(date + "T00:00:00");
  next.setDate(next.getDate() + days);
  return next.toISOString().slice(0, 10);
};
const monday = computed(() => {
  const base = new Date(today() + "T00:00:00");
  const day = base.getDay() || 7;
  base.setDate(base.getDate() + 1 - day);
  return base.toISOString().slice(0, 10);
});
const weekDays = computed(() => {
  const names = ["一", "二", "三", "四", "五", "六", "日"];
  return names.map((name, index) => {
    const key = addDays(monday.value, index);
    return { key, label: `${name} ${key.slice(5).replace("-", "/")}` };
  });
});
const weekTitle = computed(() => `本週 ${monday.value} - ${addDays(monday.value, 6)}`);
const weekEvents = computed(() => events.value.filter((event) => event.eventDate >= monday.value && event.eventDate <= addDays(monday.value, 6)));
const calendarTimes = computed(() => [...new Set(["09:00", "10:00", "11:00", "14:00", "16:00", ...weekEvents.value.map((event) => event.eventTime)])].sort());
const upcomingEvents = computed(() => events.value.filter((event) => event.eventDate >= today()).slice(0, 6));

const projectDraft = ref(defaultProject());
const eventDraft = ref(defaultEvent());

const activeCount = computed(() => projects.value.filter((project) => !project.isClosed).length);
const doneCount = computed(() => projects.value.filter((project) => project.isClosed && project.installDate).length);
const lostCount = computed(() => projects.value.filter((project) => project.isClosed && !project.installDate).length);
const viewTitle = computed(() => {
  if (view.value === "done") return "已完成結案專案";
  if (view.value === "lost") return "未成案專案";
  return "進行中專案";
});
const visibleProjects = computed(() => {
  const text = query.value.toLowerCase();
  return projects.value.filter((project) => {
    if (view.value === "active" && project.isClosed) return false;
    if (view.value === "done" && (!project.isClosed || !project.installDate)) return false;
    if (view.value === "lost" && (!project.isClosed || project.installDate)) return false;
    if (view.value === "active" && stageFilter.value !== "all" && project.status !== stageFilter.value) return false;
    if (!text) return true;
    return [project.name, project.spec, project.quoteContent, project.customNeed, project.pocResult, project.latestNote].some((value) => String(value || "").toLowerCase().includes(text));
  });
});

function defaultProject() {
  return {
    name: "新北市立醫院 列印資安專案",
    client: "新北市立醫院",
    spec: "Secure Print 標準版、浮水印、列印稽核",
    quoteContent: "企業版 + 稽核模組",
    quoteDate: today(),
    quoteNote: "報價單初版，待採購確認",
    customNeed: "AD 整合、浮水印規則、掃描管控報表",
    customDate: today(),
    customNote: "已彙整客製需求清單",
    customDays: 18,
    pocDate: "",
    pocResult: "待排程測試，已完成客戶環境盤點。",
    installDate: "",
    doneDate: "",
    isClosed: false,
    status: "need",
    owner: "KC",
    editor: "KC",
    latestNote: `${today().replaceAll("-", "/")} 建立專案並完成初步盤點。`
  };
}

function defaultEvent() {
  return { id: 0, eventDate: today(), eventTime: "09:00", type: "POC", title: "客戶 POC 排程", owner: "KC", color: "#25d7ef", editor: "KC" };
}

async function api(path, options = {}) {
  const res = await fetch(path, options);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

async function loadAll() {
  try {
    hasError.value = false;
    syncText.value = "同步中...";
    const [projectData, eventData] = await Promise.all([api("/api/projects"), api("/api/events")]);
    projects.value = projectData;
    events.value = eventData;
    lastSync.value = new Date().toLocaleTimeString("zh-TW", { hour: "2-digit", minute: "2-digit" });
    syncText.value = "後端已同步";
  } catch (error) {
    hasError.value = true;
    syncText.value = "同步失敗";
    console.error(error);
  }
}

async function createProject() {
  const saved = await api("/api/projects", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(projectDraft.value) });
  projects.value.unshift(saved);
  projectDraft.value = defaultProject();
  projectDialog.value?.close();
  syncText.value = "專案已新增";
}

async function saveProject(project) {
  const saved = await api(`/api/projects/${project.id}`, { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(project) });
  replaceProject(saved);
  syncText.value = "欄位已儲存";
}

async function closeProject(project) {
  const saved = await api(`/api/projects/${project.id}/close`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ installDate: project.installDate, editor: "KC" }) });
  replaceProject(saved);
  syncText.value = project.installDate ? "已歸檔至已完成結案" : "已歸檔至未成案";
}

async function uploadFile(project, kind, event) {
  const file = event.target.files?.[0];
  if (!file) return;
  const body = new FormData();
  body.append("file", file);
  body.append("uploadedBy", "KC");
  await api(`/api/projects/${project.id}/files?kind=${kind}`, { method: "POST", body });
  await loadAll();
  event.target.value = "";
}

function replaceProject(saved) {
  const index = projects.value.findIndex((project) => project.id === saved.id);
  if (index >= 0) projects.value[index] = saved;
}

function setView(nextView, nextStage) {
  view.value = nextView;
  stageFilter.value = nextStage;
}

function countStage(stage) {
  return projects.value.filter((project) => !project.isClosed && project.status === stage).length;
}

function labelStatus(status) {
  return stages[status]?.label || status;
}

function openProjectModal() {
  projectDraft.value = defaultProject();
  projectDialog.value?.showModal();
}

function eventsAt(date, time) {
  return events.value.filter((event) => event.eventDate === date && event.eventTime === time);
}

function openEventModal(event = null) {
  eventDraft.value = event ? { ...event } : defaultEvent();
  eventDialog.value?.showModal();
}

async function saveEvent() {
  const editing = Boolean(eventDraft.value.id);
  const path = editing ? `/api/events/${eventDraft.value.id}` : "/api/events";
  const method = editing ? "PUT" : "POST";
  const saved = await api(path, { method, headers: { "Content-Type": "application/json" }, body: JSON.stringify(eventDraft.value) });
  const index = events.value.findIndex((event) => event.id === saved.id);
  if (index >= 0) events.value[index] = saved;
  else events.value.push(saved);
  eventDialog.value?.close();
  syncText.value = "行事曆已同步";
}

async function deleteEvent() {
  await api(`/api/events/${eventDraft.value.id}`, { method: "DELETE" });
  events.value = events.value.filter((event) => event.id !== eventDraft.value.id);
  eventDialog.value?.close();
  syncText.value = "行程已刪除";
}

onMounted(() => {
  loadAll();
  window.setInterval(loadAll, 15000);
});
</script>
