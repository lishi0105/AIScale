<template>
  <div>
    <h2 style="margin: 8px 0 16px;">菜单餐次</h2>
    <DictTable
      label="餐次"
      :list="DictAPI.listMealTimes"
      :create="DictAPI.createMealTime"
      :update="DictAPI.updateMealTime"
      :remove="DictAPI.deleteMealTime"
    />
  </div>
</template>
<script setup lang="ts">
import DictTable from '@/components/DictTable.vue';
import { DictAPI } from '@/api/dict'
</script>