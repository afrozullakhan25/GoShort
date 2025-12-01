<template>
  <div class="container mx-auto px-6 py-12">
    <div class="max-w-3xl mx-auto">
      <!-- Hero Section -->
      <div class="text-center mb-12">
        <h1 class="text-white text-5xl font-bold mb-4">
          Shorten Your URLs
        </h1>
        <p class="text-white/80 text-xl">
          Fast, secure, and powerful URL shortening service
        </p>
      </div>

      <!-- Main Card -->
      <div class="card">
        <!-- Alert Messages -->
        <div v-if="message.text" :class="['alert', message.type === 'success' ? 'alert-success' : 'alert-error']">
          <i :class="['fas', message.type === 'success' ? 'fa-check-circle' : 'fa-exclamation-circle', 'mr-2']"></i>
          {{ message.text }}
        </div>

        <!-- URL Input Form -->
        <form @submit.prevent="shortenURL" class="space-y-4">
          <div>
            <label class="block text-gray-700 font-semibold mb-2">
              <i class="fas fa-link mr-2"></i>Enter your long URL
            </label>
            <input
              v-model="longURL"
              type="url"
              placeholder="https://example.com/very-long-url"
              class="input"
              required
            />
          </div>

          <div>
            <label class="block text-gray-700 font-semibold mb-2">
              <i class="fas fa-tag mr-2"></i>Custom short code (optional)
            </label>
            <input
              v-model="customCode"
              type="text"
              placeholder="my-custom-link"
              class="input"
              pattern="[a-zA-Z0-9_-]+"
            />
            <p class="text-sm text-gray-500 mt-1">Only letters, numbers, dash and underscore allowed</p>
          </div>

          <button type="submit" class="btn btn-primary w-full" :disabled="loading">
            <span v-if="!loading">
              <i class="fas fa-compress-arrows-alt mr-2"></i>Shorten URL
            </span>
            <span v-else class="flex items-center justify-center">
              <span class="loading mr-2"></span>
              Creating...
            </span>
          </button>
        </form>

        <!-- Result -->
        <div v-if="shortURL" class="mt-8 p-6 bg-gradient-to-r from-green-50 to-blue-50 rounded-lg border-2 border-green-200">
          <h3 class="text-gray-800 font-bold mb-4 text-lg">
            <i class="fas fa-check-circle text-green-500 mr-2"></i>Success!
          </h3>
          
          <div class="space-y-3">
            <div>
              <label class="text-sm text-gray-600 font-medium">Short URL:</label>
              <div class="flex items-center space-x-2 mt-1">
                <input
                  :value="shortURL"
                  readonly
                  class="input flex-1 bg-white"
                />
                <button @click="copyToClipboard" class="btn btn-primary">
                  <i class="fas fa-copy"></i>
                </button>
              </div>
            </div>

            <div>
              <label class="text-sm text-gray-600 font-medium">Original URL:</label>
              <p class="text-gray-700 break-all mt-1">{{ originalURL }}</p>
            </div>

            <div class="flex space-x-4 pt-4">
              <a :href="shortURL" target="_blank" class="btn btn-primary flex-1 text-center">
                <i class="fas fa-external-link-alt mr-2"></i>Visit
              </a>
              <button @click="reset" class="btn bg-gray-200 text-gray-700 hover:bg-gray-300 flex-1">
                <i class="fas fa-plus mr-2"></i>New URL
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Features -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mt-12">
        <div class="card text-center">
          <i class="fas fa-shield-alt text-4xl text-indigo-500 mb-4"></i>
          <h3 class="font-bold text-gray-800 mb-2">Secure</h3>
          <p class="text-gray-600 text-sm">Protected against SSRF and injection attacks</p>
        </div>
        <div class="card text-center">
          <i class="fas fa-tachometer-alt text-4xl text-purple-500 mb-4"></i>
          <h3 class="font-bold text-gray-800 mb-2">Fast</h3>
          <p class="text-gray-600 text-sm">Lightning-fast redirects with Redis caching</p>
        </div>
        <div class="card text-center">
          <i class="fas fa-chart-line text-4xl text-pink-500 mb-4"></i>
          <h3 class="font-bold text-gray-800 mb-2">Analytics</h3>
          <p class="text-gray-600 text-sm">Track clicks and monitor your links</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import api from '../services/api'

export default {
  name: 'Home',
  setup() {
    const longURL = ref('')
    const customCode = ref('')
    const shortURL = ref('')
    const originalURL = ref('')
    const loading = ref(false)
    const message = ref({ text: '', type: '' })

    const shortenURL = async () => {
      loading.value = true
      message.value = { text: '', type: '' }

      try {
        const response = await api.shortenURL(longURL.value, customCode.value)
        shortURL.value = response.short_url
        originalURL.value = response.original_url
        message.value = { text: 'URL shortened successfully!', type: 'success' }
      } catch (error) {
        message.value = {
          text: error.response?.data?.error || 'Failed to shorten URL. Please try again.',
          type: 'error'
        }
      } finally {
        loading.value = false
      }
    }

    const copyToClipboard = async () => {
      try {
        await navigator.clipboard.writeText(shortURL.value)
        message.value = { text: 'Copied to clipboard!', type: 'success' }
      } catch (error) {
        message.value = { text: 'Failed to copy', type: 'error' }
      }
    }

    const reset = () => {
      longURL.value = ''
      customCode.value = ''
      shortURL.value = ''
      originalURL.value = ''
      message.value = { text: '', type: '' }
    }

    return {
      longURL,
      customCode,
      shortURL,
      originalURL,
      loading,
      message,
      shortenURL,
      copyToClipboard,
      reset
    }
  }
}
</script>

<style scoped>
.max-w-3xl {
  max-width: 48rem;
}

.grid {
  display: grid;
}

.grid-cols-1 {
  grid-template-columns: repeat(1, minmax(0, 1fr));
}

@media (min-width: 768px) {
  .md\:grid-cols-3 {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.gap-6 {
  gap: 1.5rem;
}

.space-y-4 > * + * {
  margin-top: 1rem;
}

.space-y-3 > * + * {
  margin-top: 0.75rem;
}

.space-x-2 > * + * {
  margin-left: 0.5rem;
}

.space-x-4 > * + * {
  margin-left: 1rem;
}

.flex {
  display: flex;
}

.flex-1 {
  flex: 1 1 0%;
}

.items-center {
  align-items: center;
}

.justify-center {
  justify-content: center;
}

.text-center {
  text-align: center;
}

.break-all {
  word-break: break-all;
}
</style>

