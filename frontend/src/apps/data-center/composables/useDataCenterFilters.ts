import { reactive } from 'vue'

import type { DataCenterQuery } from '../types'

export function useDataCenterFilters() {
  const filters = reactive<DataCenterQuery>({
    time_range: 'last_7_days',
    skip: 0,
    limit: 20,
  })

  return { filters }
}
