<template>
  <div>
    <h2 style="margin: 8px 0 16px;">商品规格</h2>
    <DictTable
      label="规格"
      :list="DictAPI.listSpecs"
      :create="DictAPI.createSpec"
      :update="DictAPI.updateSpec"
      :remove="DictAPI.deleteSpec"
    />
  </div>
</template>
<script setup lang="ts">
import DictTable from '@/components/DictTable.vue'
import { DictAPI } from '@/api/dict'
</script>