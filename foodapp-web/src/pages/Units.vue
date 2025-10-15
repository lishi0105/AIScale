<template>
  <div>
    <h2 style="margin: 8px 0 16px;">商品单位</h2>
    <DictTable
      label="单位"
      :list="DictAPI.listUnits"
      :create="DictAPI.createUnit"
      :update="DictAPI.updateUnit"
      :remove="DictAPI.deleteUnit"
    />
  </div>
</template>
<script setup lang="ts">
import DictTable from '@/components/DictTable.vue'
import { DictAPI } from '@/api/dict'
</script>