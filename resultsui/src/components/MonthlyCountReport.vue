<template>
    <LineChart :chartData="getData" :height="windowHeight" />
</template>
<script lang="ts">
import { defineComponent } from 'vue';
import { LineChart } from 'vue-chart-3';

export default defineComponent({
    components: { LineChart },
    data() {
        return {
            getData: {},
            windowHeight: window.innerHeight
        }
    },
    created() {
        // make labels (x-axis)
        const currentDate = new Date();
        const currentYearMonth = new Date(currentDate.getFullYear(), currentDate.getMonth());
        const chartLabels: string[] = [];
        for (
            let pointerYearMonth = new Date(currentDate.getFullYear() - 1, currentDate.getMonth());
            pointerYearMonth < currentYearMonth;
            pointerYearMonth.setMonth(pointerYearMonth.getMonth() + 1)
        ) {
            chartLabels.push(
                pointerYearMonth.getFullYear() + "-" + ("0" + (pointerYearMonth.getMonth() + 1)).slice(-2)
            );
        }
        this.getData = {
            labels: chartLabels,
        };
        fetch('http://localhost:3000/api/getMonthlyCount').then(
            response => response.json()
        ).then(data => {
            let processedDataSet: Object[] = []
            console.log(data)
            for (const key in data) {
                let monthYearDict: { [id: string] : Number; } = {};
                for (let i = 0; i < chartLabels.length; i++) {
                    monthYearDict[chartLabels[i]] = 0;
                }
                for (let i = 0; i < data[key].length; i++) {
                    monthYearDict[data[key][i].YearMonth] = data[key][i].Count;
                }
                processedDataSet.push({
                    label: key,
                    data: Object.keys(monthYearDict).map((key) => monthYearDict[key]),
                    fill: false,
                    borderColor: "#" + Math.floor(Math.random()*16777215).toString(16)
                });
            }
            console.log(processedDataSet);
            this.getData = {
                datasets: processedDataSet,
            };
        }).catch(error => {
            console.log(error)
        });
    }
})
</script>
