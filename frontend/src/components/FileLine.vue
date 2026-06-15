<template>
  <a v-if="fileUrl" class="file-chip" :href="fileUrl" target="_blank" rel="noopener noreferrer">{{ fileName }}</a>
  <span v-else class="file-chip">未上傳</span>
</template>

<script setup>
defineProps({
  fileName: { type: String, default: "" },
  fileUrl: { type: String, default: "" }
});
</script>

