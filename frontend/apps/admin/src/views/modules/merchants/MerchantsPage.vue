<template>
  <div class="grid gap-4">
    <MerchantsHeader @new-merchant="openNew" @refresh="reload" />

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between">
        <input
          v-model.trim="searchQuery"
          type="search"
          autocomplete="off"
          placeholder="搜索商户 ID…"
          class="w-full max-w-md rounded-lg border border-slate-200 px-3 py-2 text-sm placeholder:text-slate-400"
        />
        <label class="flex items-center gap-2 text-sm text-slate-600">
          <span class="text-slate-500">匹配</span>
          <span class="font-mono text-slate-900">{{ filteredMerchants.length }}</span>
          <span class="text-slate-500">条</span>
        </label>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">商户 ID</th>
              <th class="whitespace-nowrap px-4 py-3">邮箱</th>
              <th class="whitespace-nowrap px-4 py-3">AppID</th>
              <th class="whitespace-nowrap px-4 py-3">密钥</th>
              <th class="whitespace-nowrap px-4 py-3">代收余额</th>
              <th class="whitespace-nowrap px-4 py-3">可用余额</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
              <th class="sticky right-0 z-20 whitespace-nowrap bg-slate-50 px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="8" class="px-4 py-8 text-center text-slate-500">加载中...</td>
            </tr>
            <tr v-else-if="!filteredMerchants.length">
              <td colspan="8" class="px-4 py-8 text-center text-slate-500">暂无数据</td>
            </tr>
            <tr
              v-for="m in pagedMerchants"
              v-else
              :key="m.merchant_id"
              class="border-b border-slate-100 transition hover:bg-slate-50/80"
            >
              <td class="px-4 py-3 font-mono font-semibold text-slate-900">{{ m.merchant_id }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1.5">
                  <span class="min-w-0 font-mono text-slate-700">{{ m.email || '-' }}</span>
                  <button
                    type="button"
                    class="inline-flex shrink-0 cursor-pointer items-center justify-center rounded p-0.5 text-slate-400 transition-colors duration-150 hover:text-indigo-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-400/40"
                    title="复制邮箱"
                    aria-label="复制邮箱"
                    @click="copyValue(m.email, '邮箱')"
                  >
                    <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                      />
                    </svg>
                  </button>
                </div>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1.5">
                  <span class="min-w-0 font-mono text-slate-700">{{ m.app_id || '-' }}</span>
                  <button
                    type="button"
                    class="inline-flex shrink-0 cursor-pointer items-center justify-center rounded p-0.5 text-slate-400 transition-colors duration-150 hover:text-indigo-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-400/40"
                    title="复制 AppID"
                    aria-label="复制 AppID"
                    @click="copyValue(m.app_id, 'AppID')"
                  >
                    <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                      />
                    </svg>
                  </button>
                </div>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1.5">
                  <span class="min-w-0 font-mono text-slate-700">{{ maskedSecret(m.app_secret) }}</span>
                  <button
                    type="button"
                    class="inline-flex shrink-0 cursor-pointer items-center justify-center rounded p-0.5 text-slate-400 transition-colors duration-150 hover:text-indigo-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-400/40"
                    title="复制密钥"
                    aria-label="复制密钥"
                    @click="copyValue(m.app_secret, '密钥')"
                  >
                    <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                      />
                    </svg>
                  </button>
                </div>
              </td>
              <td class="px-4 py-3 tabular-nums text-slate-700">{{ formatMoney(m.payin_balance) }}</td>
              <td class="px-4 py-3 tabular-nums text-slate-700">{{ formatMoney(m.available_balance ?? 0) }}</td>
              <td class="px-4 py-3">
                <span
                  v-if="m.status === 1"
                  class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
                >
                  启用
                </span>
                <span v-else class="rounded-full bg-rose-100 px-2 py-0.5 text-xs font-semibold text-rose-700">锁定</span>
              </td>
              <td class="sticky right-0 z-10 bg-white px-4 py-3 text-right">
                <div class="flex flex-wrap items-center justify-end gap-1.5">
                  <RouterLink
                    :to="{ path: '/settlement/withdrawals', query: { merchant_id: m.merchant_id } }"
                    class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-semibold text-slate-800 hover:border-slate-300 hover:bg-slate-50"
                  >
                    提现
                  </RouterLink>
                  <RouterLink
                    :to="{ path: '/merchants/deposit', query: { merchant_id: m.merchant_id } }"
                    class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-semibold text-slate-800 hover:border-slate-300 hover:bg-slate-50"
                  >
                    充值
                  </RouterLink>
                  <button
                    type="button"
                    class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-semibold text-slate-800 hover:border-slate-300"
                    @click="quickTransfer(m)"
                  >
                    划转
                  </button>
                  <button
                    type="button"
                    class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-semibold text-slate-800 hover:border-slate-300"
                    @click="openEdit(m.merchant_id)"
                  >
                    编辑
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <AdminPaginationBar
        v-if="!loading && filteredMerchants.length"
        :total="total"
        :page="page"
        :page-size="pageSize"
        :page-count="pageCount"
        @update:page="page = $event"
        @update:page-size="pageSize = $event"
      />
    </div>

    <UiDrawer
      v-model="drawerOpen"
      :title="drawerTitle"
      subtitle="分段维护：基本资料、API、财务与代收/代付产品；保存将写入当前表单中的全部已填字段。"
      max-width-class="max-w-3xl"
    >
      <div v-if="drawerOpen" class="space-y-4">
        <div
          class="inline-flex max-w-full flex-wrap gap-0.5 rounded-lg border border-slate-200/90 bg-slate-100 p-0.5 shadow-sm"
          role="tablist"
          aria-label="商户编辑"
        >
          <button
            v-for="tab in detailTabs"
            :key="tab.key"
            type="button"
            role="tab"
            :aria-selected="rightTab === tab.key"
            class="rounded-md px-2.5 py-1.5 text-xs font-semibold transition"
            :class="
              rightTab === tab.key
                ? 'bg-white text-slate-900 shadow-sm ring-1 ring-slate-200/80'
                : 'text-slate-600 hover:text-slate-900'
            "
            @click="rightTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>

        <div v-show="rightTab === 'basic'" role="tabpanel">
          <MerchantFormCard
            section="basic"
            v-model="form"
            :is-new="false"
            :merchant-info="selectedMerchant"
            embedded
            hide-footer-actions
            :saving="saving"
            :saved="saved"
            :error="formError"
            :can-save="canSaveForm"
            :status-for-lock="form.status"
            @save="saveForm"
            @reset="resetForm"
            @toggle-lock="toggleLock"
            @reset-secret="resetSecret"
            @reset-password="resetPassword"
          />
        </div>

        <div v-show="rightTab === 'api'" role="tabpanel">
          <MerchantFormCard
            section="api"
            v-model="form"
            :is-new="false"
            :merchant-info="selectedMerchant"
            embedded
            hide-footer-actions
            :saving="saving"
            :saved="saved"
            :error="formError"
            :can-save="canSaveForm"
            :status-for-lock="form.status"
            @save="saveForm"
            @reset="resetForm"
            @toggle-lock="toggleLock"
            @reset-secret="resetSecret"
            @reset-password="resetPassword"
          />
        </div>

        <div v-show="rightTab === 'finance'" role="tabpanel">
          <MerchantFormCard
            section="finance"
            v-model="form"
            :is-new="false"
            :merchant-info="selectedMerchant"
            embedded
            hide-footer-actions
            :saving="saving"
            :saved="saved"
            :error="formError"
            :can-save="canSaveForm"
            :status-for-lock="form.status"
            @save="saveForm"
            @reset="resetForm"
            @toggle-lock="toggleLock"
            @reset-secret="resetSecret"
            @reset-password="resetPassword"
          />
        </div>

        <div v-show="rightTab === 'bindings_payin'" role="tabpanel">
          <MerchantPayinProductsCard
            v-if="selectedMerchant"
            embedded
            :grants="selectedMerchant.payin_grants || []"
            :catalog="payinProducts"
            :loading="loadingProducts"
            :saving="bindingSaving"
            :bind-error="bindError"
            @remove="removePayinProduct"
            @add="addPayinProduct"
            @update="updatePayinGrant"
          />
          <p
            v-else
            class="rounded-lg border border-dashed border-slate-200 bg-slate-50/30 px-3 py-6 text-center text-[11px] text-slate-500"
          >
            请先保存商户基本信息后再配置代收产品。
          </p>
        </div>

        <div v-show="rightTab === 'bindings_payout'" role="tabpanel">
          <MerchantPayoutProductsCard
            v-if="selectedMerchant"
            embedded
            :grants="selectedMerchant.payout_grants || []"
            :catalog="payoutProducts"
            :loading="loadingPayoutProducts"
            :saving="bindingSaving"
            :bind-error="bindError"
            @remove="removePayoutProduct"
            @add="addPayoutProduct"
            @update="updatePayoutGrant"
          />
          <p
            v-else
            class="rounded-lg border border-dashed border-slate-200 bg-slate-50/30 px-3 py-6 text-center text-[11px] text-slate-500"
          >
            请先保存商户基本信息后再配置代付产品。
          </p>
        </div>
      </div>

      <template #footer>
        <div class="flex flex-wrap items-center justify-start gap-3">
          <template v-if="rightTab === 'basic' || rightTab === 'api' || rightTab === 'finance'">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700"
              @click="resetForm"
            >
              重置
            </button>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="saving || !canSaveForm"
              @click="saveForm"
            >
              {{ saving ? '保存中...' : '保存配置' }}
            </button>
          </template>
          <button
            type="button"
            class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700"
            @click="closeDrawer"
          >
            关闭
          </button>
        </div>
      </template>
    </UiDrawer>

    <div v-if="createWizardOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
      <div class="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-2xl border border-slate-200 bg-white shadow-xl">
        <div class="border-b border-slate-200 px-5 py-4">
          <div class="text-base font-semibold text-slate-900">
            {{ createWizardPhase === 'done' ? '商户已创建' : '新建商户' }}
          </div>
          <div class="mt-1 text-xs text-slate-500">
            <template v-if="createWizardPhase === 'done'">
              请妥善保存以下凭据（关闭后 API Secret 与密码不会在界面完整展示）。
            </template>
            <template v-else>
              {{ createWizardPhaseSubtitle }}
            </template>
          </div>
          <div
            v-if="createWizardPhase !== 'done'"
            class="mt-2 flex flex-wrap gap-1.5 text-[10px] font-medium text-slate-500"
          >
            <span :class="createPhaseChipClass(1)">1 邮箱</span>
            <span class="text-slate-300">→</span>
            <span :class="createPhaseChipClass(2)">2 财务</span>
            <span class="text-slate-300">→</span>
            <span :class="createPhaseChipClass(3)">3 代收</span>
            <span class="text-slate-300">→</span>
            <span :class="createPhaseChipClass(4)">4 代付</span>
          </div>
        </div>

        <!-- 1 邮箱 -->
        <div v-if="createWizardPhase === 1" class="space-y-4 px-5 py-4">
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">登录邮箱</span>
            <input
              v-model.trim="createEmail"
              type="email"
              autocomplete="off"
              placeholder="name@example.com"
              class="rounded-lg border border-slate-200 px-3 py-2 text-sm"
              @keydown.enter.prevent="checkCreateEmail"
            />
          </label>
          <p v-if="emailCheckHint" class="text-xs" :class="emailCheckHintClass">{{ emailCheckHint }}</p>
          <p v-if="createWizardError" class="text-xs text-rose-700">{{ createWizardError }}</p>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
              @click="closeCreateWizard"
            >
              取消
            </button>
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 disabled:opacity-40"
              :disabled="emailCheckSubmitting || !createEmail.trim()"
              @click="checkCreateEmail"
            >
              {{ emailCheckSubmitting ? '检查中...' : '检查邮箱可用' }}
            </button>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="!canGoNextFromEmailStep"
              @click="goCreateWizardPhase(2)"
            >
              下一步
            </button>
          </div>
        </div>

        <!-- 2 USDT -->
        <div v-else-if="createWizardPhase === 2" class="space-y-4 px-5 py-4">
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">提现 USDT 收款地址</span>
            <input
              v-model.trim="createWithdrawUsdt"
              type="text"
              autocomplete="off"
              class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
              placeholder="如 TRC20 地址"
            />
          </label>
          <p class="text-[11px] text-amber-800/90">请与商户确认主网；错误地址可能导致资金损失。</p>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
              @click="goCreateWizardPhase(1)"
            >
              上一步
            </button>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="!canGoNextFromFinanceStep"
              @click="goCreateWizardPhase(3)"
            >
              下一步
            </button>
          </div>
        </div>

        <!-- 3 代收 -->
        <div v-else-if="createWizardPhase === 3" class="space-y-4 px-5 py-4">
          <p class="text-[11px] text-slate-500">选择开放给该商户的代收产品；费率可在创建后在商户编辑中调整。</p>
          <div v-if="loadingProducts" class="py-6 text-center text-xs text-slate-500">加载产品中…</div>
          <template v-else>
            <div v-if="createPayinPicks.length" class="max-h-48 space-y-2 overflow-y-auto">
              <div
                v-for="row in createPayinPicks"
                :key="row.id"
                class="flex items-center justify-between gap-2 rounded-lg border border-slate-200 bg-slate-50/50 px-3 py-2 text-xs"
              >
                <span class="font-mono text-slate-800">{{ row.code }} — {{ row.name }}</span>
                <button
                  type="button"
                  class="shrink-0 text-rose-700 hover:underline"
                  @click="removeCreatePayin(row.id)"
                >
                  移除
                </button>
              </div>
            </div>
            <div v-else class="rounded-lg border border-dashed border-slate-200 py-6 text-center text-[11px] text-slate-500">
              未选择代收产品
            </div>
            <div class="flex flex-col gap-2 sm:flex-row sm:items-end">
              <label class="min-w-0 flex-1 grid gap-0.5">
                <span class="text-[11px] font-medium text-slate-600">添加代收产品</span>
                <select v-model.number="createPayinPick" class="rounded-lg border border-slate-200 px-2 py-1.5 text-sm">
                  <option :value="0">请选择</option>
                  <option v-for="p in createPayinCatalogLeft" :key="p.id" :value="p.id">{{ p.code }} — {{ p.name }}</option>
                </select>
              </label>
              <button
                type="button"
                class="shrink-0 rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
                :disabled="createPayinPick <= 0"
                @click="addCreatePayin"
              >
                添加
              </button>
            </div>
          </template>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
              @click="goCreateWizardPhase(2)"
            >
              上一步
            </button>
            <button type="button" class="rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs text-slate-700" @click="skipCreatePayin">
              跳过
            </button>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white"
              @click="goCreateWizardPhase(4)"
            >
              下一步
            </button>
          </div>
        </div>

        <!-- 4 代付 -->
        <div v-else-if="createWizardPhase === 4" class="space-y-4 px-5 py-4">
          <p class="text-[11px] text-slate-500">选择开放给该商户的代付产品；费率与计费模式可在创建后编辑。</p>
          <div v-if="loadingPayoutProducts" class="py-6 text-center text-xs text-slate-500">加载产品中…</div>
          <template v-else>
            <div v-if="createPayoutPicks.length" class="max-h-48 space-y-2 overflow-y-auto">
              <div
                v-for="row in createPayoutPicks"
                :key="row.id"
                class="flex items-center justify-between gap-2 rounded-lg border border-slate-200 bg-slate-50/50 px-3 py-2 text-xs"
              >
                <span class="font-mono text-slate-800">{{ row.code }} — {{ row.name }}</span>
                <button
                  type="button"
                  class="shrink-0 text-rose-700 hover:underline"
                  @click="removeCreatePayout(row.id)"
                >
                  移除
                </button>
              </div>
            </div>
            <div v-else class="rounded-lg border border-dashed border-slate-200 py-6 text-center text-[11px] text-slate-500">
              未选择代付产品
            </div>
            <div class="flex flex-col gap-2 sm:flex-row sm:items-end">
              <label class="min-w-0 flex-1 grid gap-0.5">
                <span class="text-[11px] font-medium text-slate-600">添加代付产品</span>
                <select v-model.number="createPayoutPick" class="rounded-lg border border-slate-200 px-2 py-1.5 text-sm">
                  <option :value="0">请选择</option>
                  <option v-for="p in createPayoutCatalogLeft" :key="p.id" :value="p.id">{{ p.code }} — {{ p.name }}</option>
                </select>
              </label>
              <button
                type="button"
                class="shrink-0 rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
                :disabled="createPayoutPick <= 0"
                @click="addCreatePayout"
              >
                添加
              </button>
            </div>
          </template>
          <p v-if="createWizardError" class="text-xs text-rose-700">{{ createWizardError }}</p>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
              @click="goCreateWizardPhase(3)"
            >
              上一步
            </button>
            <button
              type="button"
              class="rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs text-slate-700 disabled:opacity-40"
              :disabled="createSubmitting || !canSubmitCreate"
              @click="submitCreateMerchant"
            >
              跳过并创建
            </button>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="createSubmitting || !canSubmitCreate"
              @click="submitCreateMerchant"
            >
              {{ createSubmitting ? '创建中...' : '创建商户' }}
            </button>
          </div>
        </div>

        <div v-else-if="createWizardPhase === 'done'" class="space-y-3 px-5 py-4">
          <p class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-[11px] text-amber-900">
            建议立即将凭据发送给商户或保存至密管；可逐项复制或使用「一键复制全部」。
          </p>
          <div class="space-y-2">
            <div
              v-for="row in createCredentialRows"
              :key="row.key"
              class="rounded-lg border border-slate-200 bg-slate-50/50 px-3 py-2"
            >
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0 flex-1">
                  <div class="text-[11px] font-medium text-slate-600">{{ row.label }}</div>
                  <div class="mt-0.5 break-all font-mono text-xs text-slate-900">{{ row.value || '—' }}</div>
                </div>
                <button
                  type="button"
                  class="shrink-0 rounded-md border border-slate-200 bg-white px-2 py-1 text-[10px] font-semibold text-slate-700 hover:bg-slate-50"
                  @click="copyValue(row.value, row.label)"
                >
                  复制
                </button>
              </div>
            </div>
          </div>
          <div class="flex flex-wrap gap-2 border-t border-slate-100 pt-3">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
              @click="copyAllCreateCredentials"
            >
              一键复制全部
            </button>
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white"
              @click="finishCreateWizard"
            >
              完成
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="transferDialogOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
      <div class="w-full max-w-md rounded-2xl border border-slate-200 bg-white shadow-xl">
        <div class="border-b border-slate-200 px-5 py-4">
          <div class="text-base font-semibold text-slate-900">余额划转（代收 -> 代付）</div>
          <div class="mt-1 text-xs text-slate-500">商户：{{ transferTargetMerchant?.merchant_id || '-' }}</div>
        </div>
        <div class="space-y-3 px-5 py-4">
          <div class="rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm text-slate-700">
            <div>当前代收：{{ formatMoney(transferTargetMerchant?.payin_balance ?? 0) }}</div>
            <div class="mt-1">当前可用：{{ formatMoney(transferTargetMerchant?.available_balance ?? 0) }}</div>
          </div>
          <label class="grid gap-1">
            <span class="text-xs text-slate-500">划转金额（{{ transferCurrencyCode }}）</span>
            <input v-model.number="transferAmount" type="number" min="1" step="1" class="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <p v-if="transferMsg" class="text-xs text-slate-600">{{ transferMsg }}</p>
        </div>
        <div class="flex justify-end gap-2 border-t border-slate-200 px-5 py-4">
          <button
            type="button"
            class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700"
            :disabled="transferLoading"
            @click="closeTransferDialog"
          >
            取消
          </button>
          <button
            type="button"
            class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
            :disabled="transferLoading || transferAmount <= 0 || !transferTargetMerchant || transferInsufficient"
            @click="submitTransfer"
          >
            {{ transferLoading ? '划转中...' : '确认划转' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'

import AdminPaginationBar from '../../../components/AdminPaginationBar.vue'
import { UiDrawer } from '../../../../../../shared/ui'
import { useUiDialog } from '../../../composables/useUiDialog'
import { useUiToast } from '../../../composables/useUiToast'
import { useClientPagination } from '../../../composables/useClientPagination'
import { adminGet, adminPost, adminPut } from '../../../lib/adminApi'
import { adminDisplaySettings, formatAdminMoney } from '../../../lib/displaySettings'

import MerchantFormCard from './MerchantFormCard.vue'
import MerchantPayinProductsCard from './MerchantPayinProductsCard.vue'
import MerchantPayoutProductsCard from './MerchantPayoutProductsCard.vue'
import MerchantsHeader from './MerchantsHeader.vue'
import type { AdminMerchantInfo, MerchantForm, MerchantPayinGrant, MerchantPayoutGrant, ProductRow } from './types'
import { emptyMerchantForm, merchantToForm } from './types'

const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const toast = useUiToast()
const dialog = useUiDialog()

const loading = ref(false)
const loadingProducts = ref(false)
const loadingPayoutProducts = ref(false)
const saving = ref(false)
const bindingSaving = ref(false)
const saved = ref(false)
const formError = ref('')
const bindError = ref('')
const transferDialogOpen = ref(false)
const transferTargetMerchantId = ref<string | null>(null)
const transferAmount = ref(0)
const transferLoading = ref(false)
const transferMsg = ref('')

const createWizardOpen = ref(false)
type CreateWizardPhase = 1 | 2 | 3 | 4 | 'done'
const createWizardPhase = ref<CreateWizardPhase>(1)
const createWithdrawUsdt = ref('')
const createPayinPicks = ref<ProductRow[]>([])
const createPayoutPicks = ref<ProductRow[]>([])
const createPayinPick = ref(0)
const createPayoutPick = ref(0)
const createEmail = ref('')
type EmailCheckStatus = 'idle' | 'checking' | 'available' | 'taken' | 'invalid' | 'error'
const emailCheckStatus = ref<EmailCheckStatus>('idle')
const emailCheckSubmitting = ref(false)
const emailCheckErrorDetail = ref('')
const createSubmitting = ref(false)
const createWizardError = ref('')
const createDonePayload = ref<{ merchant: AdminMerchantInfo; generated_password?: string } | null>(null)

const merchants = ref<AdminMerchantInfo[]>([])
const payinProducts = ref<ProductRow[]>([])
const payoutProducts = ref<ProductRow[]>([])
const selectedMerchantId = ref<string | null>(null)
type DetailTabKey = 'basic' | 'api' | 'finance' | 'bindings_payin' | 'bindings_payout'

const rightTab = ref<DetailTabKey>('basic')
const drawerOpen = ref(false)
const searchQuery = ref('')

const detailTabs: { key: DetailTabKey; label: string }[] = [
  { key: 'basic', label: '基本资料' },
  { key: 'api', label: 'API 对接' },
  { key: 'finance', label: '财务' },
  { key: 'bindings_payin', label: '代收产品' },
  { key: 'bindings_payout', label: '代付产品' },
]

const form = ref<MerchantForm>(emptyMerchantForm())

const selectedMerchant = computed(() => {
  const id = selectedMerchantId.value
  if (!id) return null
  return merchants.value.find((m) => m.merchant_id === id) ?? null
})
const transferTargetMerchant = computed(() => {
  const id = transferTargetMerchantId.value
  if (!id) return null
  return merchants.value.find((m) => m.merchant_id === id) ?? null
})
const transferInsufficient = computed(() => {
  const m = transferTargetMerchant.value
  if (!m) return true
  const amountCent = Math.floor(transferAmount.value) * 100
  return amountCent <= 0 || amountCent > (m.payin_balance ?? 0)
})

const drawerTitle = computed(() => `编辑商户 · ${form.value.merchant_id || ''}`)

const canSaveForm = computed(() => true)

const emailCheckHint = computed(() => {
  switch (emailCheckStatus.value) {
    case 'available':
      return '该邮箱可用于创建新商户。'
    case 'taken':
      return '该邮箱已注册商户，请更换邮箱。'
    case 'invalid':
      return '请输入有效的邮箱地址。'
    case 'error':
      return emailCheckErrorDetail.value || '检查失败，请重试。'
    default:
      return ''
  }
})

const emailCheckHintClass = computed(() => {
  if (emailCheckStatus.value === 'available') return 'text-emerald-700'
  if (
    emailCheckStatus.value === 'taken' ||
    emailCheckStatus.value === 'invalid' ||
    emailCheckStatus.value === 'error'
  ) {
    return 'text-rose-700'
  }
  return 'text-slate-600'
})

const canGoNextFromEmailStep = computed(
  () => emailCheckStatus.value === 'available' && createEmail.value.includes('@'),
)

const canGoNextFromFinanceStep = computed(() => createWithdrawUsdt.value.trim().length > 0)

const canSubmitCreate = computed(
  () => emailCheckStatus.value === 'available' && createEmail.value.includes('@'),
)

const createWizardPhaseSubtitle = computed(() => {
  switch (createWizardPhase.value) {
    case 1:
      return '第 1 步：检查邮箱是否已被占用。'
    case 2:
      return '第 2 步：配置提现 USDT 链上收款地址。'
    case 3:
      return '第 3 步：绑定代收产品；可跳过。'
    case 4:
      return '第 4 步：绑定代付产品；可跳过。'
    default:
      return ''
  }
})

const createPayinCatalogLeft = computed(() => {
  const ids = new Set(createPayinPicks.value.map((p) => p.id))
  return payinProducts.value.filter((p) => !ids.has(p.id))
})

const createPayoutCatalogLeft = computed(() => {
  const ids = new Set(createPayoutPicks.value.map((p) => p.id))
  return payoutProducts.value.filter((p) => !ids.has(p.id))
})

function createPhaseChipClass(phase: number) {
  const cur = createWizardPhase.value
  const active = cur !== 'done' && cur === phase
  return active
    ? 'rounded-md bg-slate-900 px-2 py-0.5 text-white'
    : 'rounded-md bg-slate-100 px-2 py-0.5 text-slate-600'
}

const createCredentialRows = computed(() => {
  const p = createDonePayload.value
  if (!p) return [] as { key: string; label: string; value: string }[]
  const m = p.merchant
  return [
    { key: 'merchant_id', label: '商户 ID', value: m.merchant_id },
    { key: 'email', label: '登录邮箱', value: m.email || '' },
    { key: 'app_id', label: 'AppID', value: m.app_id || '' },
    { key: 'app_secret', label: 'API Secret', value: m.app_secret || '' },
    { key: 'password', label: '登录密码', value: p.generated_password || '' },
  ]
})
const transferCurrencyCode = computed(() => adminDisplaySettings.value.currency_code || 'CNY')

const filteredMerchants = computed(() => {
  const s = searchQuery.value.trim().toLowerCase()
  const list = merchants.value
  if (!s) return list
  return list.filter((m) =>
    (m.merchant_id || '').toLowerCase().includes(s) ||
    (m.email || '').toLowerCase().includes(s) ||
    (m.app_id || '').toLowerCase().includes(s),
  )
})
function maskedSecret(secret: string) {
  const s = String(secret || '')
  if (s.length <= 8) return s || '-'
  return `${s.slice(0, 4)}****${s.slice(-4)}`
}

async function copyValue(value: string, label: string) {
  const text = String(value || '').trim()
  if (!text) {
    toast.error(`${label}为空，无法复制`)
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    toast.success(`${label}已复制`)
  } catch {
    toast.error(`复制${label}失败`)
  }
}


const { page, pageSize, total, pageCount, slice: pagedMerchants } = useClientPagination(filteredMerchants, 10)

watch(searchQuery, () => {
  page.value = 1
})

function formatMoney(v: number) {
  return formatAdminMoney(v)
}

function normalizedPayoutGrants(m: AdminMerchantInfo): MerchantPayoutGrant[] {
  const grants = m.payout_grants || []
  if (grants.length > 0) {
    return grants.map((g) => ({
      payout_product_id: g.payout_product_id,
      merchant_rate_bps: g.merchant_rate_bps ?? 0,
      fee_mode: g.fee_mode || 1,
      fee_fixed_amount: g.fee_fixed_amount ?? 0,
    }))
  }
  return (m.payout_product_ids || []).map((id) => ({
    payout_product_id: id,
    merchant_rate_bps: 0,
    fee_mode: 1,
    fee_fixed_amount: 0,
  }))
}

function normalizedPayinGrants(m: AdminMerchantInfo): MerchantPayinGrant[] {
  const grants = m.payin_grants || []
  if (grants.length > 0) {
    return grants.map((g) => ({
      payin_product_id: g.payin_product_id,
      merchant_rate_bps: g.merchant_rate_bps ?? 0,
    }))
  }
  return (m.payin_product_ids || []).map((id) => ({
    payin_product_id: id,
    merchant_rate_bps: 0,
  }))
}

function applySelectedToForm() {
  const m = selectedMerchant.value
  if (!m) return
  form.value = merchantToForm(m)
}

function resetForm() {
  saved.value = false
  formError.value = ''
  applySelectedToForm()
}

function openEdit(merchantId: string) {
  selectedMerchantId.value = merchantId
  rightTab.value = 'basic'
  applySelectedToForm()
  saved.value = false
  formError.value = ''
  drawerOpen.value = true
}

const emailLooksValid = (s: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s.trim())

function openNew() {
  createWizardPhase.value = 1
  createWithdrawUsdt.value = ''
  createPayinPicks.value = []
  createPayoutPicks.value = []
  createPayinPick.value = 0
  createPayoutPick.value = 0
  createEmail.value = ''
  emailCheckStatus.value = 'idle'
  emailCheckErrorDetail.value = ''
  createWizardError.value = ''
  createDonePayload.value = null
  createWizardOpen.value = true
  void loadPayinProducts()
  void loadPayoutProducts()
}

function goCreateWizardPhase(phase: 1 | 2 | 3 | 4) {
  createWizardPhase.value = phase
}

function skipCreatePayin() {
  createPayinPicks.value = []
  createPayinPick.value = 0
  goCreateWizardPhase(4)
}

function addCreatePayin() {
  const id = createPayinPick.value
  if (id <= 0) return
  const p = payinProducts.value.find((x) => x.id === id)
  if (!p || createPayinPicks.value.some((x) => x.id === id)) return
  createPayinPicks.value = [...createPayinPicks.value, p]
  createPayinPick.value = 0
}

function removeCreatePayin(id: number) {
  createPayinPicks.value = createPayinPicks.value.filter((x) => x.id !== id)
}

function addCreatePayout() {
  const id = createPayoutPick.value
  if (id <= 0) return
  const p = payoutProducts.value.find((x) => x.id === id)
  if (!p || createPayoutPicks.value.some((x) => x.id === id)) return
  createPayoutPicks.value = [...createPayoutPicks.value, p]
  createPayoutPick.value = 0
}

function removeCreatePayout(id: number) {
  createPayoutPicks.value = createPayoutPicks.value.filter((x) => x.id !== id)
}

function closeCreateWizard() {
  createWizardOpen.value = false
}

async function checkCreateEmail() {
  const raw = createEmail.value.trim()
  createWizardError.value = ''
  emailCheckErrorDetail.value = ''
  if (!raw) {
    emailCheckStatus.value = 'invalid'
    return
  }
  if (!emailLooksValid(raw)) {
    emailCheckStatus.value = 'invalid'
    return
  }
  emailCheckSubmitting.value = true
  emailCheckStatus.value = 'checking'
  try {
    const q = new URLSearchParams({ email: raw.toLowerCase() })
    const res = await adminGet<{ available: boolean }>(`/v1/admin/merchants/email_available?${q.toString()}`)
    emailCheckStatus.value = res.available ? 'available' : 'taken'
  } catch (e) {
    emailCheckStatus.value = 'error'
    emailCheckErrorDetail.value = e instanceof Error ? e.message : String(e)
  } finally {
    emailCheckSubmitting.value = false
  }
}

async function submitCreateMerchant() {
  if (createWizardPhase.value !== 4) return
  if (!canSubmitCreate.value) return
  createWizardError.value = ''
  createSubmitting.value = true
  try {
    const resp = await adminPost<{ merchant: AdminMerchantInfo; generated_password?: string }>('/v1/admin/merchants', {
      email: createEmail.value.trim().toLowerCase(),
      withdraw_usdt_address: createWithdrawUsdt.value.trim(),
      payin_product_ids: createPayinPicks.value.map((p) => p.id),
      payout_product_ids: createPayoutPicks.value.map((p) => p.id),
    })
    createDonePayload.value = {
      merchant: resp.merchant,
      generated_password: resp.generated_password,
    }
    createWizardPhase.value = 'done'
    toast.success('商户创建成功')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    createWizardError.value = msg
    toast.error(`创建失败：${msg}`)
  } finally {
    createSubmitting.value = false
  }
}

async function copyAllCreateCredentials() {
  const lines = createCredentialRows.value.map((r) => `${r.label}：${r.value}`)
  const text = lines.join('\n')
  if (!text.trim()) {
    toast.error('没有可复制的内容')
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    toast.success('全部凭据已复制')
  } catch {
    toast.error('复制失败')
  }
}

async function finishCreateWizard() {
  createWizardOpen.value = false
  await reload()
}

watch(createEmail, () => {
  if (emailCheckStatus.value !== 'idle') {
    emailCheckStatus.value = 'idle'
    emailCheckErrorDetail.value = ''
  }
})

async function loadPayinProducts() {
  loadingProducts.value = true
  try {
    const res = await adminGet<{ products: ProductRow[] }>('/v1/admin/payin_products')
    payinProducts.value = res.products || []
  } catch (e) {
    payinProducts.value = []
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`加载代收产品失败：${msg}`)
  } finally {
    loadingProducts.value = false
  }
}

async function loadPayoutProducts() {
  loadingPayoutProducts.value = true
  try {
    const res = await adminGet<{ products: ProductRow[] }>('/v1/admin/payout_products')
    payoutProducts.value = res.products || []
  } catch (e) {
    payoutProducts.value = []
    const msg = e instanceof Error ? e.message : String(e)
    toast.error(`加载代付产品失败：${msg}`)
  } finally {
    loadingPayoutProducts.value = false
  }
}

async function reload() {
  formError.value = ''
  saved.value = false
  loading.value = true
  try {
    const data = await adminGet<{ merchants: AdminMerchantInfo[] }>('/v1/admin/merchants')
    merchants.value = data.merchants || []
    if (selectedMerchantId.value && merchants.value.some((m) => m.merchant_id === selectedMerchantId.value)) {
      applySelectedToForm()
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    formError.value = msg
    toast.error(`加载商户列表失败：${msg}`)
  } finally {
    loading.value = false
  }
}

async function saveForm() {
  const mid = selectedMerchant.value?.merchant_id
  if (!mid) return
  saving.value = true
  formError.value = ''
  saved.value = false
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(mid)}`, {
      status: form.value.status,
      notify_url: form.value.notify_url,
      return_url: form.value.return_url,
      ip_whitelist: form.value.ip_whitelist,
      withdraw_usdt_address: form.value.withdraw_usdt_address.trim(),
      payin_grants: normalizedPayinGrants(selectedMerchant.value!),
      payout_grants: normalizedPayoutGrants(selectedMerchant.value!),
    })
    const row = resp.merchant
    const idx = merchants.value.findIndex((m) => m.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
    saved.value = true
    closeDrawer()
    toast.success('编辑已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    formError.value = msg
    toast.error(`保存商户失败：${msg}`)
  } finally {
    saving.value = false
  }
}

async function toggleLock() {
  const m = selectedMerchant.value
  if (!m) return
  formError.value = ''
  try {
    const target = m.status === 1 ? 0 : 1
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`, {
      status: target,
      notify_url: form.value.notify_url,
      return_url: form.value.return_url,
      ip_whitelist: form.value.ip_whitelist,
      withdraw_usdt_address: form.value.withdraw_usdt_address.trim() || '',
      payin_grants: normalizedPayinGrants(m),
      payout_grants: normalizedPayoutGrants(m),
    })
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
    saved.value = true
    toast.success('状态已更新')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    formError.value = msg
    toast.error(`更新商户状态失败：${msg}`)
  }
}

async function resetSecret() {
  const m = selectedMerchant.value
  if (!m) return
  const ok = await dialog.confirm('确认重置该商户密钥？重置后旧密钥立即失效。', '重置密钥')
  if (!ok) return
  formError.value = ''
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(
      `/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`,
      {
        reset_secret: true,
        status: form.value.status,
        notify_url: form.value.notify_url,
        return_url: form.value.return_url,
        ip_whitelist: form.value.ip_whitelist,
        withdraw_usdt_address: form.value.withdraw_usdt_address.trim() || '',
        payin_grants: normalizedPayinGrants(m),
        payout_grants: normalizedPayoutGrants(m),
      },
    )
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
    saved.value = true
    toast.success('密钥已重置')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    formError.value = msg
    toast.error(`重置密钥失败：${msg}`)
  }
}

async function resetPassword() {
  const m = selectedMerchant.value
  if (!m) return
  const ok = await dialog.confirm('确认重置该商户登录密码？重置后旧密码立即失效。', '重置密码')
  if (!ok) return
  formError.value = ''
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo; generated_password?: string }>(
      `/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`,
      {
        reset_password: true,
        status: form.value.status,
        notify_url: form.value.notify_url,
        return_url: form.value.return_url,
        ip_whitelist: form.value.ip_whitelist,
        withdraw_usdt_address: form.value.withdraw_usdt_address.trim() || '',
        payin_grants: normalizedPayinGrants(m),
        payout_grants: normalizedPayoutGrants(m),
      },
    )
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
    saved.value = true
    if (resp.generated_password) {
      await copyValue(resp.generated_password, '新密码')
      toast.success(`密码已重置：${resp.generated_password}`)
    } else {
      toast.success('密码已重置')
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    formError.value = msg
    toast.error(`重置密码失败：${msg}`)
  }
}

async function persistPayinProducts(newGrants: MerchantPayinGrant[]) {
  const m = selectedMerchant.value
  if (!m) return
  bindingSaving.value = true
  bindError.value = ''
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`, {
      status: form.value.status,
      notify_url: form.value.notify_url,
      return_url: form.value.return_url,
      ip_whitelist: form.value.ip_whitelist,
      withdraw_usdt_address: form.value.withdraw_usdt_address.trim() || '',
      payin_grants: newGrants,
      payout_grants: normalizedPayoutGrants(m),
    })
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
    toast.success('代收产品授权已更新')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    bindError.value = msg
    toast.error(`保存代收产品授权失败：${msg}`)
  } finally {
    bindingSaving.value = false
  }
}

async function persistPayoutProducts(newGrants: MerchantPayoutGrant[]) {
  const m = selectedMerchant.value
  if (!m) return
  bindingSaving.value = true
  bindError.value = ''
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`, {
      status: form.value.status,
      notify_url: form.value.notify_url,
      return_url: form.value.return_url,
      ip_whitelist: form.value.ip_whitelist,
      withdraw_usdt_address: form.value.withdraw_usdt_address.trim() || '',
      payin_grants: normalizedPayinGrants(m),
      payout_grants: newGrants,
    })
    const row = resp.merchant
    const idx = merchants.value.findIndex((x) => x.merchant_id === row.merchant_id)
    if (idx >= 0) merchants.value[idx] = row
    form.value = merchantToForm(row)
    toast.success('代付产品授权已更新')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    bindError.value = msg
    toast.error(`保存代付产品授权失败：${msg}`)
  } finally {
    bindingSaving.value = false
  }
}

async function removePayinProduct(productId: number) {
  const m = selectedMerchant.value
  if (!m) return
  const ok = await dialog.confirm('确认移除该代收产品绑定？', '移除代收产品')
  if (!ok) return
  const cur = normalizedPayinGrants(m)
  const next = cur.filter((x) => x.payin_product_id !== productId)
  void persistPayinProducts(next)
}

function addPayinProduct(productId: number) {
  const m = selectedMerchant.value
  if (!m || productId <= 0) return
  const cur = normalizedPayinGrants(m)
  if (cur.some((x) => x.payin_product_id === productId)) return
  cur.push({
    payin_product_id: productId,
    merchant_rate_bps: 0,
  })
  void persistPayinProducts(cur)
}

function updatePayinGrant(grant: MerchantPayinGrant) {
  const m = selectedMerchant.value
  if (!m) return
  const cur = normalizedPayinGrants(m)
  const idx = cur.findIndex((x) => x.payin_product_id === grant.payin_product_id)
  if (idx < 0) return
  cur[idx] = {
    payin_product_id: grant.payin_product_id,
    merchant_rate_bps: grant.merchant_rate_bps ?? 0,
  }
  void persistPayinProducts(cur)
}

async function removePayoutProduct(productId: number) {
  const m = selectedMerchant.value
  if (!m) return
  const ok = await dialog.confirm('确认移除该代付产品绑定？', '移除代付产品')
  if (!ok) return
  const cur = normalizedPayoutGrants(m)
  const next = cur.filter((x) => x.payout_product_id !== productId)
  void persistPayoutProducts(next)
}

function addPayoutProduct(productId: number) {
  const m = selectedMerchant.value
  if (!m || productId <= 0) return
  const cur = normalizedPayoutGrants(m)
  if (cur.some((x) => x.payout_product_id === productId)) return
  cur.push({
    payout_product_id: productId,
    merchant_rate_bps: 0,
    fee_mode: 1,
    fee_fixed_amount: 0,
  })
  void persistPayoutProducts(cur)
}

function updatePayoutGrant(grant: MerchantPayoutGrant) {
  const m = selectedMerchant.value
  if (!m) return
  const cur = normalizedPayoutGrants(m)
  const idx = cur.findIndex((x) => x.payout_product_id === grant.payout_product_id)
  if (idx < 0) return
  cur[idx] = {
    payout_product_id: grant.payout_product_id,
    merchant_rate_bps: grant.merchant_rate_bps ?? 0,
    fee_mode: grant.fee_mode || 1,
    fee_fixed_amount: grant.fee_fixed_amount ?? 0,
  }
  void persistPayoutProducts(cur)
}

function closeDrawer() {
  drawerOpen.value = false
}

function quickTransfer(m: AdminMerchantInfo) {
  transferTargetMerchantId.value = m.merchant_id
  transferAmount.value = 0
  transferMsg.value = ''
  transferDialogOpen.value = true
}

function closeTransferDialog() {
  transferDialogOpen.value = false
  transferTargetMerchantId.value = null
  transferAmount.value = 0
  transferMsg.value = ''
}

async function submitTransfer() {
  const m = transferTargetMerchant.value
  if (!m || transferAmount.value <= 0) return
  const amountCent = Math.floor(transferAmount.value) * 100
  if (amountCent > (m.payin_balance ?? 0)) {
    transferMsg.value = '划转金额不能超过当前代收余额'
    toast.error('划转金额超过代收余额')
    return
  }
  transferLoading.value = true
  transferMsg.value = ''
  try {
    const resp = await adminPost<{ ok: boolean; payin_balance: number; available_balance: number }>(
      `/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}/transfer_payin_to_payout`,
      { amount: amountCent, reason: 'ADMIN_QUICK_TRANSFER' },
    )
    m.payin_balance = resp.payin_balance
    m.available_balance = resp.available_balance
    transferMsg.value = `划转成功：代收 ${formatMoney(resp.payin_balance)}，可用 ${formatMoney(resp.available_balance)}`
    transferAmount.value = 0
    toast.success('划转成功')
    closeTransferDialog()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    transferMsg.value = msg
    toast.error(`余额划转失败：${msg}`)
  } finally {
    transferLoading.value = false
  }
}

watch(drawerOpen, (open, wasOpen) => {
  if (wasOpen === true && open === false) void reload()
})

let unregister: (() => void) | null = null
onMounted(() => {
  void loadPayinProducts()
  void loadPayoutProducts()
  void reload()
  if (registerRefresh) unregister = registerRefresh(() => void reload())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>
