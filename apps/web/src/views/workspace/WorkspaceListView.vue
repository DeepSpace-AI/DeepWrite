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
import IconViewGridOutline from "~icons/mdi/view-grid-outline";
import IconViewListOutline from "~icons/mdi/view-list-outline";
import IconAccountGroupOutline from "~icons/mdi/account-group-outline";
import IconFileDocumentOutline from "~icons/mdi/file-document-outline";
import IconDotsVertical from "~icons/mdi/dots-vertical";
import IconPlus from "~icons/mdi/plus";
import IconPencilOutline from "~icons/mdi/pencil-outline";
import IconDeleteOutline from "~icons/mdi/delete-outline";

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
const pageSize = ref(9);

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
  <section class="scholar-workspace">
    <!-- Hero Header Section -->
    <header class="scholar-hero px-8 pt-10 pb-8">
      <div class="mx-auto max-w-6xl">
        <div class="flex flex-wrap items-start justify-between gap-6">
          <div>
            <p class="label-sm text-[var(--color-on-surface-variant)] uppercase tracking-widest">
              Workspace / Directory
            </p>
            <h1 class="text-editorial mt-3 text-4xl font-light text-[var(--color-on-background)]">
              {{ t("workspace.title") }}
            </h1>
            <p class="body-md mt-3 text-[var(--color-on-surface-variant)]">
              {{ t("workspace.subtitle") }}
            </p>
          </div>
          <button
            type="button"
            class="scholar-btn-primary flex items-center gap-2 rounded-lg px-5 py-2.5"
            @click="openCreateModal"
          >
            <IconPlus class="h-5 w-5" />
            {{ t("workspace.createBtn") }}
          </button>
        </div>
      </div>
    </header>

    <!-- Toolbar Section -->
    <div class="scholar-toolbar px-8 py-5">
      <div class="mx-auto flex max-w-6xl flex-wrap items-center gap-4">
        <div class="scholar-search relative flex-1 max-w-md">
          <IconSearch class="scholar-search-icon absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2" />
          <input
            v-model="searchKeyword"
            type="text"
            class="scholar-input w-full rounded-lg py-3 pl-12 pr-4"
            :placeholder="t('workspace.searchPlaceholder')"
            @input="applySearch"
          />
        </div>

        <div class="scholar-view-toggle flex items-center gap-1 rounded-lg p-1">
          <button
            type="button"
            class="scholar-view-btn rounded-md p-2"
            :class="viewMode === 'card' ? 'active' : ''"
            @click="switchView('card')"
          >
            <IconViewGridOutline class="h-5 w-5" />
          </button>
          <button
            type="button"
            class="scholar-view-btn rounded-md p-2"
            :class="viewMode === 'table' ? 'active' : ''"
            @click="switchView('table')"
          >
            <IconViewListOutline class="h-5 w-5" />
          </button>
        </div>

        <!-- Results Info -->
        <p class="label-sm text-[var(--color-on-surface-variant)]">
          {{ t("workspace.resultsInfo", { total: filteredWorkspaces.length, current: currentPage, pages: totalPages }) }}
        </p>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="scholar-content px-8 pb-8">
      <div class="mx-auto max-w-6xl">
        <!-- Error Messages -->
        <div v-if="fetchError" class="scholar-error mb-6 rounded-lg p-4">
          {{ fetchError }}
        </div>
        <div v-if="actionError" class="scholar-error mb-6 rounded-lg p-4">
          {{ actionError }}
        </div>

        <!-- Loading State -->
        <section v-if="isLoading" class="scholar-grid-container rounded-xl p-6">
          <div class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
            <div v-for="skeleton in 6" :key="`skeleton-${skeleton}`" class="skeleton h-48 rounded-xl" />
          </div>
        </section>

        <!-- Card View -->
        <section v-else-if="viewMode === 'card'" class="scholar-grid-container rounded-xl p-6">
          <div class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
            <article
              v-for="item in pagedWorkspaces"
              :key="item.id"
              class="scholar-card group rounded-xl p-6 transition-all"
            >
              <!-- Header -->
              <div class="flex items-start justify-between gap-3">
                <h3 class="text-editorial text-lg font-medium text-[var(--color-on-background)]">
                  {{ item.name }}
                </h3>
                <span
                  class="scholar-status label-sm rounded-full px-2 py-0.5"
                  :class="item.status === 'active' ? 'status-active' : 'status-archived'"
                >
                  {{ item.status === "active" ? t("workspace.statusActive") : t("workspace.statusArchived") }}
                </span>
              </div>

              <!-- Stats -->
              <div class="mt-4 flex items-center gap-6">
                <div class="flex items-center gap-2 text-sm text-[var(--color-on-surface-variant)]">
                  <IconAccountGroupOutline class="h-4 w-4" />
                  <span>{{ item.members }}</span>
                </div>
                <div class="flex items-center gap-2 text-sm text-[var(--color-on-surface-variant)]">
                  <IconFileDocumentOutline class="h-4 w-4" />
                  <span>{{ item.docs }}</span>
                </div>
              </div>

              <!-- Footer -->
              <div class="mt-6 flex items-center justify-between text-xs text-[var(--color-on-surface-muted)]">
                <span class="scholar-role">{{ roleLabelMap[item.role] }}</span>
                <span>{{ t("workspace.updatedAt", { date: item.updatedAt }) }}</span>
              </div>

              <!-- Actions -->
              <div class="mt-4 flex items-center justify-end gap-2 pt-4">
                <button
                  type="button"
                  class="scholar-btn-enter label-sm rounded-lg px-4 py-2"
                  @click="enterWorkspace(item)"
                >
                  {{ t("workspace.enterBtn") }}
                </button>
                <button
                  type="button"
                  class="scholar-btn-more rounded-lg p-2"
                  :data-more-trigger-id="item.id"
                  @click="toggleMoreMenu($event, item.id)"
                >
                  <IconDotsVertical class="h-5 w-5" />
                </button>
              </div>
            </article>

            <!-- Empty State -->
            <div v-if="!pagedWorkspaces.length" class="scholar-empty col-span-full rounded-xl p-12 text-center">
              <p class="text-[var(--color-on-surface-variant)]">{{ t("workspace.noResults") }}</p>
            </div>
          </div>
        </section>

        <!-- Table View -->
        <section v-else class="scholar-table-container overflow-hidden rounded-xl">
          <table class="scholar-table w-full">
            <thead>
              <tr>
                <th class="text-left">{{ t("workspace.colName") }}</th>
                <th class="text-left">{{ t("workspace.colRole") }}</th>
                <th class="text-left">{{ t("workspace.colMembers") }}</th>
                <th class="text-left">{{ t("workspace.colDocs") }}</th>
                <th class="text-left">{{ t("workspace.colStatus") }}</th>
                <th class="text-left">{{ t("workspace.colUpdatedAt") }}</th>
                <th class="text-right">{{ t("workspace.colActions") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in pagedWorkspaces" :key="item.id">
                <td class="text-editorial font-medium">{{ item.name }}</td>
                <td>{{ roleLabelMap[item.role] }}</td>
                <td>{{ item.members }}</td>
                <td>{{ item.docs }}</td>
                <td>
                  <span
                    class="scholar-status label-sm rounded-full px-2 py-0.5"
                    :class="item.status === 'active' ? 'status-active' : 'status-archived'"
                  >
                    {{ item.status === "active" ? t("workspace.statusActive") : t("workspace.statusArchived") }}
                  </span>
                </td>
                <td class="text-[var(--color-on-surface-muted)]">{{ item.updatedAt }}</td>
                <td class="text-right">
                  <button
                    type="button"
                    class="scholar-btn-enter label-sm rounded-lg px-3 py-1.5"
                    @click="enterWorkspace(item)"
                  >
                    {{ t("workspace.enterBtn") }}
                  </button>
                  <button
                    type="button"
                    class="scholar-btn-more ml-2 rounded-lg p-1.5"
                    :data-more-trigger-id="item.id"
                    @click="toggleMoreMenu($event, item.id)"
                  >
                    <IconDotsVertical class="h-4 w-4" />
                  </button>
                </td>
              </tr>
              <tr v-if="!pagedWorkspaces.length">
                <td colspan="7" class="py-12 text-center text-[var(--color-on-surface-variant)]">
                  {{ t("workspace.noResults") }}
                </td>
              </tr>
            </tbody>
          </table>
        </section>

        <!-- Pagination -->
        <nav class="scholar-pagination mt-6 flex items-center justify-center gap-2 rounded-xl py-4">
          <button
            type="button"
            class="scholar-page-btn rounded-lg px-3 py-2"
            :disabled="currentPage <= 1"
            @click="prevPage"
          >
            {{ t("workspace.prevPage") }}
          </button>

          <button
            v-for="page in pageNumbers"
            :key="page"
            type="button"
            class="scholar-page-btn rounded-lg px-3 py-2"
            :class="page === currentPage ? 'active' : ''"
            @click="goToPage(page)"
          >
            {{ page }}
          </button>

          <button
            type="button"
            class="scholar-page-btn rounded-lg px-3 py-2"
            :disabled="currentPage >= totalPages"
            @click="nextPage"
          >
            {{ t("workspace.nextPage") }}
          </button>
        </nav>
      </div>
    </div>

    <!-- Modals -->
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

    <!-- More Menu -->
    <Teleport to="body">
      <div
        v-if="openedWorkspaceItem && openedMoreMenuId"
        ref="moreMenuPanelRef"
        class="scholar-menu rounded-lg p-2"
        :style="moreMenuStyle"
      >
        <button
          type="button"
          class="scholar-menu-item flex w-full items-center gap-3 rounded-lg px-3 py-2"
          :disabled="actionLoadingId === openedWorkspaceItem.id"
          @click="openRenameModal(openedWorkspaceItem)"
        >
          <IconPencilOutline class="h-4 w-4" />
          {{ t("workspace.rename") }}
        </button>
        <button
          type="button"
          class="scholar-menu-item danger flex w-full items-center gap-3 rounded-lg px-3 py-2"
          :disabled="actionLoadingId === openedWorkspaceItem.id"
          @click="removeWorkspace(openedWorkspaceItem)"
        >
          <IconDeleteOutline class="h-4 w-4" />
          {{ t("workspace.delete") }}
        </button>
      </div>
    </Teleport>
  </section>
</template>

<style scoped>
.scholar-workspace {
  background-color: var(--surface);
  min-height: calc(100vh - 4rem);
}

.scholar-hero {
  background-color: var(--surface-container-low);
}

.scholar-toolbar {
  background-color: var(--surface-container);
}

.scholar-content {
  background-color: var(--surface);
}

.scholar-btn-primary {
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dim));
  color: var(--color-on-primary);
  font-family: var(--font-sans);
  font-weight: 500;
  transition: opacity 0.3s ease-out;
}

.scholar-btn-primary:hover {
  opacity: 0.88;
}

.scholar-search {
  position: relative;
}

.scholar-search-icon {
  color: var(--color-on-surface-variant);
  opacity: 0.5;
}

.scholar-input {
  background-color: var(--surface-container-lowest);
  color: var(--color-on-surface);
  border: none;
  outline: none;
  transition: background-color 0.3s;
}

.scholar-input:focus {
  background-color: var(--surface-container-lowest);
}

.scholar-input::placeholder {
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.scholar-view-toggle {
  background-color: var(--surface-container-low);
}

.scholar-view-btn {
  color: var(--color-on-surface-variant);
  transition: background-color 0.2s, color 0.2s;
}

.scholar-view-btn:hover {
  background-color: var(--surface-container-high);
}

.scholar-view-btn.active {
  background-color: var(--surface-container-lowest);
  color: var(--color-on-surface);
}

.scholar-grid-container {
  background-color: var(--surface-container);
}

.scholar-card {
  background-color: var(--surface-container-lowest);
  box-shadow: 0 4px 24px oklch(0.28 0.008 105 / 0.06);
}

.scholar-card:hover {
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.10);
  transform: translateY(-2px);
}

[data-theme="vellum-dark"] .scholar-card {
  box-shadow: 0 4px 24px oklch(0.15 0.02 75 / 0.15);
}

[data-theme="vellum-dark"] .scholar-card:hover {
  box-shadow: 0 8px 32px oklch(0.15 0.02 75 / 0.20);
}

.scholar-status {
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.scholar-status.status-active {
  background-color: var(--color-secondary-container);
  color: var(--color-on-secondary-container);
}

.scholar-status.status-archived {
  background-color: var(--surface-container-high);
  color: var(--color-on-surface-variant);
}

.scholar-btn-enter {
  background-color: var(--color-primary);
  color: var(--color-on-primary);
  font-weight: 500;
  transition: opacity 0.2s;
}

.scholar-btn-enter:hover {
  opacity: 0.88;
}

.scholar-btn-more {
  color: var(--color-on-surface-variant);
  transition: background-color 0.2s;
}

.scholar-btn-more:hover {
  background-color: var(--surface-container-high);
}

.scholar-error {
  background-color: var(--color-error-container);
  color: var(--color-on-error-container);
}

.scholar-empty {
  background-color: var(--surface-container-low);
}

.scholar-table-container {
  background-color: var(--surface-container-lowest);
  box-shadow: 0 4px 24px oklch(0.28 0.008 105 / 0.06);
}

.scholar-table thead {
  background-color: var(--surface-container-low);
}

.scholar-table th {
  padding: 1rem;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-on-surface-variant);
}

.scholar-table td {
  padding: 1rem;
  color: var(--color-on-surface);
  border-top: 1px solid var(--surface-container);
}

.scholar-table tr:hover td {
  background-color: var(--surface-container-low);
}

.scholar-pagination {
  background-color: var(--surface-container-low);
}

.scholar-page-btn {
  background-color: transparent;
  color: var(--color-on-surface-variant);
  font-size: 0.875rem;
  transition: background-color 0.2s, color 0.2s;
}

.scholar-page-btn:hover:not(:disabled) {
  background-color: var(--surface-container);
}

.scholar-page-btn.active {
  background-color: var(--color-primary);
  color: var(--color-on-primary);
}

.scholar-page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.scholar-menu {
  background-color: var(--surface-raised);
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.15);
}

[data-theme="vellum-dark"] .scholar-menu {
  box-shadow: 0 8px 32px oklch(0.15 0.02 75 / 0.25);
}

.scholar-menu-item {
  font-size: 0.875rem;
  color: var(--color-on-surface);
  transition: background-color 0.2s;
}

.scholar-menu-item:hover {
  background-color: var(--surface-container);
}

.scholar-menu-item.danger {
  color: var(--color-error);
}

.scholar-menu-item.danger:hover {
  background-color: var(--color-error-container);
}
</style>