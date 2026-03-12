<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { ApiError } from "@/api/http";
import {
  createWorkspace,
  deleteWorkspace,
  listWorkspaces,
  updateWorkspace,
  type Workspace,
} from "@/api/workspace";
import { useUserStore } from "@/stores/user";
import CreateWorkspaceModal from "@/views/workspace/modal/CreateWorkspaceModal.vue";
import RenameWorkspaceModal from "@/views/workspace/modal/RenameWorkspaceModal.vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import IconSearch from "~icons/mdi/magnify";

type WorkspaceRole = "owner" | "admin" | "editor" | "viewer";
type ViewMode = "card" | "table";

interface WorkspaceItem {
  id: string;
  name: string;
  role: WorkspaceRole;
  members: number;
  docs: number;
  updatedAt: string;
  status: "active" | "archived";
}

const userStore = useUserStore();
const router = useRouter();
const { t, locale } = useI18n();

const rawWorkspaces = ref<Workspace[]>([]);
const isLoading = ref(false);
const fetchError = ref("");

const viewMode = ref<ViewMode>("card");
const searchKeyword = ref("");
const currentPage = ref(1);
const pageSize = ref(6);

const isCreateModalOpen = ref(false);
const isSubmittingCreate = ref(false);
const createErrorMessage = ref("");
const isRenameModalOpen = ref(false);
const isSubmittingRename = ref(false);
const renameErrorMessage = ref("");
const renameTarget = ref<WorkspaceItem | null>(null);
const actionError = ref("");
const actionLoadingId = ref("");
const openedMoreMenuId = ref("");
const moreMenuStyle = ref<Record<string, string>>({});
const moreMenuPanelRef = ref<HTMLElement | null>(null);

const roleLabelMap = computed<Record<WorkspaceRole, string>>(() => ({
  owner: t("workspace.roleOwner"),
  admin: t("workspace.roleAdmin"),
  editor: t("workspace.roleEditor"),
  viewer: t("workspace.roleViewer"),
}));

function normalizeRole(value: string): WorkspaceRole {
  if (value === "owner" || value === "admin" || value === "editor" || value === "viewer") {
    return value;
  }
  return "viewer";
}

function formatUpdatedAt(value: string): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  const dateLocale = locale.value === "zh-CN" ? "zh-CN" : "en-US";
  return d.toLocaleString(dateLocale, { hour12: false });
}

const workspaceItems = computed<WorkspaceItem[]>(() => {
  const currentUserId = userStore.user?.id || "";

  return rawWorkspaces.value.map((item) => {
    const currentMember = item.members?.find((m) => m.user_id === currentUserId);
    const role = normalizeRole(
      currentMember?.role || (item.owner_id === currentUserId ? "owner" : "viewer"),
    );

    return {
      id: item.id,
      name: item.name,
      role,
      members: item.members?.length || 1,
      docs: 0,
      updatedAt: formatUpdatedAt(item.updated_at),
      status: item.status === "archived" ? "archived" : "active",
    };
  });
});

const filteredWorkspaces = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase();
  if (!keyword) return workspaceItems.value;

  return workspaceItems.value.filter((item) => {
    const roleLabel = roleLabelMap.value[item.role];
    return (
      item.name.toLowerCase().includes(keyword) ||
      roleLabel.toLowerCase().includes(keyword) ||
      item.status.toLowerCase().includes(keyword)
    );
  });
});

const totalPages = computed(() => {
  const pages = Math.ceil(filteredWorkspaces.value.length / pageSize.value);
  return pages > 0 ? pages : 1;
});

const pagedWorkspaces = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return filteredWorkspaces.value.slice(start, start + pageSize.value);
});

const pageNumbers = computed(() => Array.from({ length: totalPages.value }, (_, i) => i + 1));

const openedWorkspaceItem = computed(() => {
  if (!openedMoreMenuId.value) return null;
  return workspaceItems.value.find((item) => item.id === openedMoreMenuId.value) || null;
});

async function loadWorkspaces() {
  isLoading.value = true;
  fetchError.value = "";

  try {
    rawWorkspaces.value = await listWorkspaces(100, 0);
  } catch (error) {
    fetchError.value = error instanceof ApiError ? error.message : t("workspace.loadError");
  } finally {
    isLoading.value = false;
  }
}

function switchView(mode: ViewMode) {
  closeMoreMenu();
  viewMode.value = mode;
}

function applySearch() {
  closeMoreMenu();
  currentPage.value = 1;
}

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return;
  closeMoreMenu();
  currentPage.value = page;
}

function prevPage() {
  goToPage(currentPage.value - 1);
}

function nextPage() {
  goToPage(currentPage.value + 1);
}

function openCreateModal() {
  closeMoreMenu();
  createErrorMessage.value = "";
  isCreateModalOpen.value = true;
}

function openRenameModal(item: WorkspaceItem) {
  closeMoreMenu();
  renameTarget.value = item;
  renameErrorMessage.value = "";
  isRenameModalOpen.value = true;
}

function closeMoreMenu() {
  openedMoreMenuId.value = "";
}

function openMoreMenu(event: MouseEvent, workspaceId: string) {
  const trigger = event.currentTarget as HTMLElement | null;
  if (!trigger) return;

  const rect = trigger.getBoundingClientRect();
  const menuWidth = 160;
  const menuTop = rect.bottom + 8 + window.scrollY;
  const minLeft = window.scrollX + 8;
  const maxLeft = window.scrollX + window.innerWidth - menuWidth - 8;
  const menuLeft = Math.min(Math.max(rect.right + window.scrollX - menuWidth, minLeft), maxLeft);

  openedMoreMenuId.value = workspaceId;
  moreMenuStyle.value = {
    position: "absolute",
    top: `${menuTop}px`,
    left: `${menuLeft}px`,
    width: "160px",
    zIndex: "1000",
  };
}

function toggleMoreMenu(event: MouseEvent, workspaceId: string) {
  if (openedMoreMenuId.value === workspaceId) {
    closeMoreMenu();
    return;
  }
  openMoreMenu(event, workspaceId);
}

function handleDocumentPointerDown(event: PointerEvent) {
  if (!openedMoreMenuId.value) return;

  const target = event.target as Node | null;
  if (!target) return;

  if (moreMenuPanelRef.value?.contains(target)) return;

  const triggerSelector = `[data-more-trigger-id="${openedMoreMenuId.value}"]`;
  if (target instanceof Element && target.closest(triggerSelector)) return;

  closeMoreMenu();
}

function handleWindowChange() {
  closeMoreMenu();
}

async function handleCreateWorkspace(payload: {
  name: string;
  description: string;
  public: boolean;
}) {
  if (isSubmittingCreate.value) return;

  isSubmittingCreate.value = true;
  createErrorMessage.value = "";

  try {
    const created = await createWorkspace(payload);
    rawWorkspaces.value = [created, ...rawWorkspaces.value];
    isCreateModalOpen.value = false;
    currentPage.value = 1;
  } catch (error) {
    createErrorMessage.value =
      error instanceof ApiError ? error.message : t("workspace.createError");
  } finally {
    isSubmittingCreate.value = false;
  }
}

async function handleRenameWorkspace(payload: { name: string }) {
  if (isSubmittingRename.value || !renameTarget.value) return;

  const currentTarget = renameTarget.value;
  const trimmedName = payload.name.trim();
  if (!trimmedName || trimmedName === currentTarget.name) {
    isRenameModalOpen.value = false;
    return;
  }

  isSubmittingRename.value = true;
  renameErrorMessage.value = "";

  try {
    await updateWorkspace(currentTarget.id, { name: trimmedName });
    await loadWorkspaces();
    isRenameModalOpen.value = false;
    renameTarget.value = null;
  } catch (error) {
    renameErrorMessage.value =
      error instanceof ApiError ? error.message : t("workspace.renameError");
  } finally {
    isSubmittingRename.value = false;
  }
}

function enterWorkspace(item: WorkspaceItem) {
  closeMoreMenu();
  router.push({ name: "workspace-detail", params: { id: item.id } });
}

async function removeWorkspace(item: WorkspaceItem) {
  closeMoreMenu();
  const confirmed = window.confirm(t("workspace.deleteConfirm", { name: item.name }));
  if (!confirmed) return;

  actionError.value = "";
  actionLoadingId.value = item.id;
  try {
    await deleteWorkspace(item.id);
    rawWorkspaces.value = rawWorkspaces.value.filter((w) => w.id !== item.id);
    if (currentPage.value > totalPages.value) {
      currentPage.value = totalPages.value;
    }
  } catch (error) {
    actionError.value = error instanceof ApiError ? error.message : t("workspace.deleteError");
  } finally {
    actionLoadingId.value = "";
  }
}

onMounted(() => {
  document.addEventListener("pointerdown", handleDocumentPointerDown);
  window.addEventListener("resize", handleWindowChange);
  window.addEventListener("scroll", handleWindowChange, true);
  loadWorkspaces();
});

onBeforeUnmount(() => {
  document.removeEventListener("pointerdown", handleDocumentPointerDown);
  window.removeEventListener("resize", handleWindowChange);
  window.removeEventListener("scroll", handleWindowChange, true);
});
</script>

<template>
  <section class="space-y-5">
    <section class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="text-[11px] font-mono uppercase tracking-[0.2em] text-base-content/45">
            Workspace / Directory
          </p>
          <h2 class="heading-serif mt-2 text-3xl font-bold text-base-content">
            {{ t("workspace.title") }}
          </h2>
          <p class="mt-2 max-w-2xl text-sm leading-7 text-base-content/62">
            {{ t("workspace.subtitle") }}
          </p>
        </div>
        <button type="button" class="btn btn-primary rounded-sm" @click="openCreateModal">
          {{ t("workspace.createBtn") }}
        </button>
      </div>

      <div class="mt-5 flex flex-wrap items-center gap-3">
        <label class="input input-bordered flex w-full max-w-md items-center gap-2 rounded-sm">
          <IconSearch class="h-4 w-4 text-base-content/50" />
          <input
            v-model="searchKeyword"
            type="text"
            class="grow"
            :placeholder="t('workspace.searchPlaceholder')"
            @input="applySearch"
          />
        </label>

        <div class="tabs tabs-box rounded-sm border border-base-300 bg-base-200 p-1">
          <button
            type="button"
            class="tab rounded-sm"
            :class="viewMode === 'card' ? 'tab-active' : ''"
            @click="switchView('card')"
          >
            {{ t("workspace.viewCard") }}
          </button>
          <button
            type="button"
            class="tab rounded-sm"
            :class="viewMode === 'table' ? 'tab-active' : ''"
            @click="switchView('table')"
          >
            {{ t("workspace.viewTable") }}
          </button>
        </div>
      </div>

      <div class="mt-3 text-xs text-base-content/50">
        {{
          t("workspace.resultsInfo", {
            total: filteredWorkspaces.length,
            current: currentPage,
            pages: totalPages,
          })
        }}
      </div>
    </section>

    <p
      v-if="fetchError"
      class="rounded-sm border border-error/30 bg-error/10 px-4 py-3 text-sm text-error"
    >
      {{ fetchError }}
    </p>

    <p
      v-if="actionError"
      class="rounded-sm border border-error/30 bg-error/10 px-4 py-3 text-sm text-error"
    >
      {{ actionError }}
    </p>

    <section v-if="isLoading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <div
        v-for="skeleton in 3"
        :key="`skeleton-${skeleton}`"
        class="h-36 animate-pulse rounded-sm border border-base-300 bg-base-200/60"
      />
    </section>

    <section v-else-if="viewMode === 'card'" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <article
        v-for="item in pagedWorkspaces"
        :key="item.id"
        class="rounded-sm border border-base-300 bg-base-100 p-4 shadow-sm"
      >
        <div class="flex items-center justify-between gap-3">
          <h3 class="truncate text-base font-semibold text-base-content">{{ item.name }}</h3>
          <span
            class="badge rounded-sm"
            :class="item.status === 'active' ? 'badge-success badge-outline' : 'badge-ghost'"
          >
            {{
              item.status === "active" ? t("workspace.statusActive") : t("workspace.statusArchived")
            }}
          </span>
        </div>

        <div class="mt-3 grid grid-cols-2 gap-2 text-xs">
          <div class="rounded-sm bg-base-200/70 px-3 py-2 text-base-content/70">
            {{ t("workspace.memberCount", { count: item.members }) }}
          </div>
          <div class="rounded-sm bg-base-200/70 px-3 py-2 text-base-content/70">
            {{ t("workspace.docCount", { count: item.docs }) }}
          </div>
        </div>

        <div class="mt-3 flex items-center justify-between text-xs text-base-content/55">
          <span>{{ roleLabelMap[item.role] }}</span>
          <span>{{ t("workspace.updatedAt", { date: item.updatedAt }) }}</span>
        </div>

        <div class="mt-4 flex items-center justify-end gap-2">
          <button type="button" class="btn btn-sm rounded-sm" @click="enterWorkspace(item)">
            {{ t("workspace.enterBtn") }}
          </button>
          <button
            type="button"
            class="btn btn-sm btn-ghost rounded-sm"
            :data-more-trigger-id="item.id"
            @click="toggleMoreMenu($event, item.id)"
          >
            {{ t("workspace.moreBtn") }}
          </button>
        </div>
      </article>

      <p
        v-if="!pagedWorkspaces.length"
        class="col-span-full rounded-sm border border-base-300 bg-base-100 px-4 py-10 text-center text-sm text-base-content/60"
      >
        {{ t("workspace.noResults") }}
      </p>
    </section>

    <section v-else-if="!isLoading" class="rounded-sm border border-base-300 bg-base-100 shadow-sm">
      <div class="border-b border-base-300 px-4 py-3">
        <h3 class="text-sm font-semibold text-base-content">{{ t("workspace.tableTitle") }}</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="table table-zebra">
          <thead>
            <tr>
              <th>{{ t("workspace.colName") }}</th>
              <th>{{ t("workspace.colRole") }}</th>
              <th>{{ t("workspace.colMembers") }}</th>
              <th>{{ t("workspace.colDocs") }}</th>
              <th>{{ t("workspace.colStatus") }}</th>
              <th>{{ t("workspace.colUpdatedAt") }}</th>
              <th class="text-right">{{ t("workspace.colActions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in pagedWorkspaces" :key="`${item.id}-row`">
              <td class="font-medium text-base-content">{{ item.name }}</td>
              <td>{{ roleLabelMap[item.role] }}</td>
              <td>{{ item.members }}</td>
              <td>{{ item.docs }}</td>
              <td>
                <span
                  class="badge rounded-sm"
                  :class="item.status === 'active' ? 'badge-success badge-outline' : 'badge-ghost'"
                >
                  {{
                    item.status === "active"
                      ? t("workspace.statusActive")
                      : t("workspace.statusArchived")
                  }}
                </span>
              </td>
              <td>{{ item.updatedAt }}</td>
              <td class="relative">
                <div class="flex items-center justify-end gap-2">
                  <button type="button" class="btn btn-xs rounded-sm" @click="enterWorkspace(item)">
                    {{ t("workspace.enterBtn") }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-xs btn-ghost rounded-sm"
                    :data-more-trigger-id="item.id"
                    @click="toggleMoreMenu($event, item.id)"
                  >
                    {{ t("workspace.moreBtn") }}
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!pagedWorkspaces.length">
              <td colspan="7" class="py-10 text-center text-sm text-base-content/60">
                {{ t("workspace.noResults") }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="flex flex-wrap items-center justify-end gap-2">
      <button
        type="button"
        class="btn btn-sm rounded-sm"
        :disabled="currentPage <= 1"
        @click="prevPage"
      >
        {{ t("workspace.prevPage") }}
      </button>

      <button
        v-for="page in pageNumbers"
        :key="`page-${page}`"
        type="button"
        class="btn btn-sm rounded-sm"
        :class="page === currentPage ? 'btn-primary' : 'btn-ghost'"
        @click="goToPage(page)"
      >
        {{ page }}
      </button>

      <button
        type="button"
        class="btn btn-sm rounded-sm"
        :disabled="currentPage >= totalPages"
        @click="nextPage"
      >
        {{ t("workspace.nextPage") }}
      </button>
    </section>

    <CreateWorkspaceModal
      v-model="isCreateModalOpen"
      :submitting="isSubmittingCreate"
      :error-message="createErrorMessage"
      @submit="handleCreateWorkspace"
    />

    <RenameWorkspaceModal
      v-model="isRenameModalOpen"
      :workspace-name="renameTarget?.name || ''"
      :submitting="isSubmittingRename"
      :error-message="renameErrorMessage"
      @submit="handleRenameWorkspace"
    />

    <Teleport to="body">
      <div
        v-if="openedWorkspaceItem && openedMoreMenuId"
        ref="moreMenuPanelRef"
        class="menu rounded-sm border border-base-300 bg-base-100 p-1 shadow-lg"
        :style="moreMenuStyle"
      >
        <button
          type="button"
          class="btn btn-ghost btn-sm w-full justify-start rounded-sm"
          :disabled="actionLoadingId === openedWorkspaceItem.id"
          @click="openRenameModal(openedWorkspaceItem)"
        >
          {{ t("workspace.rename") }}
        </button>
        <button
          type="button"
          class="btn btn-ghost btn-sm w-full justify-start rounded-sm text-error"
          :disabled="actionLoadingId === openedWorkspaceItem.id"
          @click="removeWorkspace(openedWorkspaceItem)"
        >
          {{ t("workspace.delete") }}
        </button>
      </div>
    </Teleport>
  </section>
</template>
