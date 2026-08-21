<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { AlertTriangle, ClipboardCheck, RefreshCw, Wrench } from 'lucide-vue-next'
import { useOperations } from './store'

const store = useOperations()
const tab = ref<'risk'|'facility'>('risk')
const openRisk = computed(() => store.risks.filter(item => item.open_incidents > 0 || item.overdue_days > 0))
onMounted(() => store.refresh())
</script>

<template>
  <div class="shell">
    <aside>
      <div class="brand"><Wrench :size="20"/> 公共设施维护台账</div>
      <button :class="{active:tab==='risk'}" @click="tab='risk'"><AlertTriangle :size="18"/> 风险总览</button>
      <button :class="{active:tab==='facility'}" @click="tab='facility'"><ClipboardCheck :size="18"/> 设施台账</button>
    </aside>
    <main>
      <header><div><h1>{{ tab==='risk' ? '风险总览' : '设施台账' }}</h1><p>市民服务中心 · 设施保障组</p></div><button class="icon" title="刷新" @click="store.refresh"><RefreshCw :size="18"/></button></header>
      <div v-if="store.error" class="error">{{ store.error }}</div>
      <section v-if="tab==='risk'">
        <div class="metrics"><div><span>需关注</span><strong>{{ openRisk.length }}</strong></div><div><span>开放异常</span><strong>{{ store.risks.reduce((n,r)=>n+r.open_incidents,0) }}</strong></div><div><span>逾期设施</span><strong>{{ store.risks.filter(r=>r.overdue_days>0).length }}</strong></div></div>
        <table><thead><tr><th>设施编号</th><th>状态</th><th>逾期</th><th>开放异常</th><th>处置提示</th></tr></thead><tbody><tr v-for="risk in store.risks" :key="risk.facility_id"><td>{{ risk.facility_id }}</td><td><span class="status">{{ risk.state }}</span></td><td>{{ risk.overdue_days }} 天</td><td>{{ risk.open_incidents }}</td><td>{{ risk.alternative_hint || '按计划维护' }}</td></tr></tbody></table>
      </section>
      <section v-else><table><thead><tr><th>设施</th><th>类别</th><th>重要级别</th><th>当前状态</th></tr></thead><tbody><tr v-for="facility in store.facilities" :key="facility.id"><td><b>{{ facility.name }}</b><small>{{ facility.id }}</small></td><td>{{ facility.category }}</td><td>{{ facility.criticality }}</td><td><span class="status">{{ facility.state }}</span></td></tr></tbody></table></section>
    </main>
  </div>
</template>
