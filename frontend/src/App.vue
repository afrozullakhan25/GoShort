<template>
  <div id="app" class="min-h-screen bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500">
    <nav class="bg-white/10 backdrop-blur-lg shadow-lg">
      <div class="container mx-auto px-6 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <i class="fas fa-link text-white text-3xl"></i>
            <h1 class="text-white text-2xl font-bold">GoShort</h1>
          </div>
          <div class="flex space-x-4">
            <router-link to="/" class="text-white hover:text-gray-200 px-4 py-2 rounded-lg transition">
              <i class="fas fa-home mr-2"></i>Home
            </router-link>
            <router-link to="/stats" class="text-white hover:text-gray-200 px-4 py-2 rounded-lg transition">
              <i class="fas fa-chart-bar mr-2"></i>Stats
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <router-view />
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

