<template>
    <div>
        <LineChart :chartData="getData" :height="chartHeight"/>
    </div>
</template>
<script lang="ts" setup>
import { Chart, registerables } from "chart.js";
import { reactive } from 'vue';
import { LineChart } from 'vue-chart-3';

// make chart fill window heightwise
const chartHeight = window.innerHeight
// make chart
Chart.register(...registerables);
let getData = reactive({ labels: new Array<String>(), datasets: {} });
getData.labels = createChartLabels();
(async function () {
    try {
        putReportOnChart(await (await fetch('http://localhost:3000/report/getmonthlycount')).json());
    } catch (error) {
        alert(error);
        console.log(error);
    }
})();

function createChartLabels() : string[] {
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
    return chartLabels;
}

function putReportOnChart(data: any) {
    let processedDataSet: Object[] = []
    // for each skill, use monthYearDictTemplate to create a new monthYearDict
    let monthYearDictTemplate: { [id: string] : Number; } = {};
    for (const dateLabelKey in getData.labels)
    {
        monthYearDictTemplate[getData.labels[dateLabelKey].toString()];
    }
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
    getData.datasets = processedDataSet;
}
</script>
