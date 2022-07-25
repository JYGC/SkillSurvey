<template>
    <LineChart :chartData="getData" :height="windowHeight" />
</template>
<script lang="ts">
// Active chart support
import { Chart, registerables } from "chart.js";
Chart.register(...registerables);

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
        // get data from API and put them on chart
        fetch('http://localhost:3000/api/getmonthlycount').then(
            response => response.json()
        ).then(data => {
            let processedDataSet: Object[] = []
            // for each skill, use monthYearDictTemplate to create a new monthYearDict
            let monthYearDictTemplate: { [id: string] : Number; } = {};
            chartLabels.forEach((el) => monthYearDictTemplate[el] = 0);
            for (const key in data) {
                // monthYearDict is used to put counts in the correct monthYear on the chart
                let monthYearDict = JSON.parse(JSON.stringify(monthYearDictTemplate));
                for (let i = 0; i < data[key].length; i++) {
                    monthYearDict[data[key][i].YearMonth] = data[key][i].Count;
                }
                processedDataSet.push({
                    label: key,
                    data: Object.keys(monthYearDict).map((key) => monthYearDict[key]),
                    fill: false,
                    borderColor: "#" + Math.floor(Math.random()*16777215).toString(16),
                    hidden: true,
                });
            }
            this.getData = {
                datasets: processedDataSet,
            };
        }).catch(error => {
            alert(error)
            console.log(error)
        });
    }
})
</script>
