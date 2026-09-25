package com.example.kalasetu.features.event

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.automirrored.filled.KeyboardArrowLeft
import androidx.compose.material.icons.automirrored.filled.KeyboardArrowRight
import androidx.compose.material.icons.filled.KeyboardArrowDown
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.datetime.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TimelineAndLocationScreen(
    onNext: (location: String, startDate: LocalDate?, endDate: LocalDate?) -> Unit,
    onBack: () -> Unit,
) {
    var location by remember { mutableStateOf("") }
    var selectedYear by remember { mutableStateOf(2025) }
    var selectedMonth by remember { mutableStateOf(9) }
    var startDate by remember { mutableStateOf<LocalDate?>(null) }
    var endDate by remember { mutableStateOf<LocalDate?>(null) }

    Scaffold(
        floatingActionButton = {
            FloatingActionButton(
                onClick = { onNext(location, startDate, endDate) },
                containerColor = LightPurpleBg,
                contentColor = TextDark,
                shape = CircleShape,
                modifier = Modifier.size(56.dp),
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                    contentDescription = "Next",
                    modifier = Modifier.size(24.dp),
                )
            }
        },
        containerColor = MaterialTheme.colorScheme.surface,
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(horizontal = 24.dp)
                .verticalScroll(rememberScrollState())
        ) {
            Spacer(modifier = Modifier.height(16.dp))

            IconButton(
                onClick = onBack,
                modifier = Modifier.offset(x = (-12).dp)
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                    tint = MaterialTheme.colorScheme.onSurface
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Select the Application\nTimeline and Location",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = TextDark,
                lineHeight = 38.sp
            )

            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = "Select when applications\nopen and close",
                fontSize = 24.sp,
                fontWeight = FontWeight.Normal,
                color = TextGray,
                lineHeight = 32.sp
            )

            Spacer(modifier = Modifier.height(32.dp))

            CalendarCard(
                selectedYear = selectedYear,
                selectedMonth = selectedMonth,
                startDate = startDate,
                endDate = endDate,
                onMonthChange = { selectedMonth = it },
                onYearChange = { selectedYear = it },
                onDateClick = { date ->
                    if (startDate == null || (startDate != null && endDate != null)) {
                        startDate = date
                        endDate = null
                    } else if (startDate != null && endDate == null) {
                        if (date < startDate!!) {
                            startDate = date
                        } else {
                            endDate = date
                        }
                    }
                }
            )

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Location",
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                color = TextDark,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            OutlinedTextField(
                value = location,
                onValueChange = { location = it },
                placeholder = {
                    Text(
                        "Enter the Location where the event is at",
                        color = FadedGray,
                        fontSize = 14.sp
                    )
                },
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(8.dp),
                colors = OutlinedTextFieldDefaults.colors(
                    unfocusedBorderColor = BorderGray,
                    focusedBorderColor = PurplePrimary,
                    unfocusedContainerColor = MaterialTheme.colorScheme.surfaceVariant,
                    focusedContainerColor = MaterialTheme.colorScheme.surfaceVariant
                ),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(100.dp))
        }
    }
}

@Composable
fun CalendarCard(
    selectedYear: Int,
    selectedMonth: Int,
    startDate: LocalDate?,
    endDate: LocalDate?,
    onMonthChange: (Int) -> Unit,
    onYearChange: (Int) -> Unit,
    onDateClick: (LocalDate) -> Unit
) {
    val months = listOf(
        "Jan", "Feb", "Mar", "Apr", "May", "Jun",
        "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"
    )

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = androidx.compose.foundation.BorderStroke(1.dp, BorderGray)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {

            // ─── Month / Year Selector Row ───
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Previous Month button
                IconButton(onClick = {
                    if (selectedMonth == 1) {
                        onMonthChange(12)
                        onYearChange(selectedYear - 1)
                    } else {
                        onMonthChange(selectedMonth - 1)
                    }
                }) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Filled.KeyboardArrowLeft,
                        contentDescription = "Previous Month",
                        tint = TextDark
                    )
                }

                // Month + Year dropdowns
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    var monthExpanded by remember { mutableStateOf(false) }
                    Box {
                        Row(
                            modifier = Modifier
                                .clip(RoundedCornerShape(8.dp))
                                .border(1.dp, BorderGray, RoundedCornerShape(8.dp))
                                .clickable { monthExpanded = true }
                                .padding(horizontal = 12.dp, vertical = 6.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(months[selectedMonth - 1], fontSize = 14.sp, color = TextDark)
                            Spacer(modifier = Modifier.width(4.dp))
                            Icon(
                                Icons.Default.KeyboardArrowDown,
                                contentDescription = null,
                                modifier = Modifier.size(16.dp),
                                tint = TextDark
                            )
                        }
                        DropdownMenu(
                            expanded = monthExpanded,
                            onDismissRequest = { monthExpanded = false }
                        ) {
                            months.forEachIndexed { index, month ->
                                DropdownMenuItem(
                                    text = { Text(month) },
                                    onClick = {
                                        onMonthChange(index + 1)
                                        monthExpanded = false
                                    }
                                )
                            }
                        }
                    }

                    var yearExpanded by remember { mutableStateOf(false) }
                    Box {
                        Row(
                            modifier = Modifier
                                .clip(RoundedCornerShape(8.dp))
                                .border(1.dp, BorderGray, RoundedCornerShape(8.dp))
                                .clickable { yearExpanded = true }
                                .padding(horizontal = 12.dp, vertical = 6.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(selectedYear.toString(), fontSize = 14.sp, color = TextDark)
                            Spacer(modifier = Modifier.width(4.dp))
                            Icon(
                                Icons.Default.KeyboardArrowDown,
                                contentDescription = null,
                                modifier = Modifier.size(16.dp),
                                tint = TextDark
                            )
                        }
                        DropdownMenu(
                            expanded = yearExpanded,
                            onDismissRequest = { yearExpanded = false }
                        ) {
                            (2024..2030).forEach { year ->
                                DropdownMenuItem(
                                    text = { Text(year.toString()) },
                                    onClick = {
                                        onYearChange(year)
                                        yearExpanded = false
                                    }
                                )
                            }
                        }
                    }
                }

                // Next Month button
                IconButton(onClick = {
                    if (selectedMonth == 12) {
                        onMonthChange(1)
                        onYearChange(selectedYear + 1)
                    } else {
                        onMonthChange(selectedMonth + 1)
                    }
                }) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Filled.KeyboardArrowRight,
                        contentDescription = "Next Month",
                        tint = TextDark
                    )
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // ─── Days of Week Header ───
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                val days = listOf("Su", "Mo", "Tu", "We", "Th", "Fr", "Sa")
                days.forEach { day ->
                    Text(
                        text = day,
                        fontSize = 12.sp,
                        color = TextGray,
                        modifier = Modifier.weight(1f),
                        textAlign = TextAlign.Center
                    )
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            // ─── Calendar Grid ───
            val currentMonth = LocalDate(selectedYear, selectedMonth, 1)
            val daysInMonth = currentMonth.daysUntil(currentMonth.plus(1, DateTimeUnit.MONTH))
            val firstDayOfWeek = currentMonth.dayOfWeek.isoDayNumber % 7

            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                for (week in 0 until 6) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        for (day in 0 until 7) {
                            val cellIndex = week * 7 + day
                            val dayOfMonth = cellIndex - firstDayOfWeek + 1

                            if (dayOfMonth in 1..daysInMonth) {
                                val date = LocalDate(selectedYear, selectedMonth, dayOfMonth)
                                val isStart = date == startDate
                                val isEnd = date == endDate
                                val isInRange = startDate != null && endDate != null &&
                                        date > startDate && date < endDate

                                CalendarDay(
                                    day = dayOfMonth.toString(),
                                    isStart = isStart,
                                    isEnd = isEnd,
                                    isInRange = isInRange,
                                    onClick = { onDateClick(date) }
                                )
                            } else if (dayOfMonth > daysInMonth) {
                                val nextMonthDate = currentMonth.plus((dayOfMonth - 1).toLong(), DateTimeUnit.DAY)
                                CalendarDay(
                                    day = nextMonthDate.dayOfMonth.toString(),
                                    isStart = false,
                                    isEnd = false,
                                    isInRange = false,
                                    isFaded = true,
                                    onClick = { }
                                )
                            } else {
                                Spacer(modifier = Modifier.weight(1f).height(32.dp))
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun RowScope.CalendarDay(
    day: String,
    isStart: Boolean,
    isEnd: Boolean,
    isInRange: Boolean,
    isFaded: Boolean = false,
    onClick: () -> Unit
) {
    val bgColor = when {
        isStart || isEnd -> PurplePrimary
        isInRange -> LightPurpleBg
        else -> Color.Transparent
    }

    val textColor = when {
        isStart || isEnd -> Color.White
        isInRange -> PurplePrimary
        isFaded -> FadedGray
        else -> TextDark
    }

    Box(
        modifier = Modifier
            .weight(1f)
            .aspectRatio(1f)
            .padding(2.dp)
            .clip(CircleShape)
            .background(bgColor)
            .clickable(enabled = !isFaded) { onClick() },
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = day,
            fontSize = 14.sp,
            color = textColor,
            fontWeight = if (isStart || isEnd || isInRange) FontWeight.Bold else FontWeight.Normal
        )
    }
}