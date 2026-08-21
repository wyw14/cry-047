import { defineStore } from 'pinia'

export type Facility = { id: string; name: string; category: string; state: string; criticality: string }
export type Risk = { facility_id: string; state: string; overdue_days: number; open_incidents: number; alternative_hint: string }

const headers = { 'X-Actor-ID': 'person-planner', 'X-Actor-Role': 'admin', 'X-Actor-Name': '值班管理员' }

export const useOperations = defineStore('operations', {
  state: () => ({ facilities: [] as Facility[], risks: [] as Risk[], busy: false, error: '' }),
  actions: {
    async refresh() {
      this.busy = true
      this.error = ''
      try {
        const [facilities, risks] = await Promise.all([
          fetch('/api/v1/facilities?sort=name', { headers }).then(assertOK),
          fetch('/api/v1/risk-board', { headers }).then(assertOK)
        ])
        this.facilities = (await facilities.json()).items
        this.risks = await risks.json()
      } catch (error) {
        this.error = error instanceof Error ? error.message : '加载失败'
      } finally {
        this.busy = false
      }
    }
  }
})

function assertOK(response: Response) {
  if (!response.ok) throw new Error(`服务返回 ${response.status}`)
  return response
}
