<template>
  <div class="container mx-auto px-6 py-12">
    <div class="max-w-4xl mx-auto">
      <h2 class="text-white text-4xl font-bold mb-8 text-center">
        <i class="fas fa-chart-bar mr-3"></i>URL Statistics
      </h2>

      <div class="card">
        <div class="mb-6">
          <input
            v-model="searchCode"
            type="text"
            placeholder="Enter short code to view stats..."
            class="input"
            @keyup.enter="fetchStats"
          />
          <button @click="fetchStats" class="btn btn-primary w-full mt-4">
            <i class="fas fa-search mr-2"></i>Search
          </button>
        </div>

        <div v-if="loading" class="text-center py-8">
          <span class="loading"></span>
          <p class="text-gray-600 mt-4">Loading...</p>
        </div>

        <div v-else-if="stats" class="space-y-6">
          <div class="grid grid-cols-2 gap-4">
            <div class="p-4 bg-indigo-50 rounded-lg">
              <div class="text-indigo-600 text-3xl font-bold">{{ stats.click_count }}</div>
              <div class="text-gray-600 text-sm mt-1">Total Clicks</div>
            </div>
            <div class="p-4 bg-purple-50 rounded-lg">
              <div class="text-purple-600 text-3xl font-bold">
                {{ stats.is_active ? 'Active' : 'Inactive' }}
              </div>
              <div class="text-gray-600 text-sm mt-1">Status</div>
            </div>
          </div>

          <div>
            <label class="text-sm text-gray-600 font-medium">Short Code:</label>
            <p class="text-gray-800 font-mono mt-1">{{ stats.short_code }}</p>
          </div>

          <div>
            <label class="text-sm text-gray-600 font-medium">Original URL:</label>
            <p class="text-gray-800 break-all mt-1">{{ stats.original_url }}</p>
          </div>

          <div>
            <label class="text-sm text-gray-600 font-medium">Created:</label>
            <p class="text-gray-800 mt-1">{{ formatDate(stats.created_at) }}</p>
          </div>
        </div>

        <div v-else-if="error" class="alert alert-error">
          <i class="fas fa-exclamation-circle mr-2"></i>{{ error }}
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import api from '../services/api'

export default {
  name: 'Stats',
  setup() {
    const searchCode = ref('')
    const stats = ref(null)
    const loading = ref(false)
    const error = ref('')

    const fetchStats = async () => {
      if (!searchCode.value) return

      loading.value = true
      error.value = ''
      stats.value = null

      try {
        stats.value = await api.getURLStats(searchCode.value)
      } catch (err) {
        error.value = err.response?.data?.error || 'URL not found'
      } finally {
        loading.value = false
      }
    }

    const formatDate = (dateString) => {
      return new Date(dateString).toLocaleString()
    }

    return {
      searchCode,
      stats,
      loading,
      error,
      fetchStats,
      formatDate
    }
  }
}
</script>

