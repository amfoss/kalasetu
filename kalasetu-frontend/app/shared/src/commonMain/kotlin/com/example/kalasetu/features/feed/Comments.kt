package com.example.kalasetu.features.feed

import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.outlined.FavoriteBorder
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil3.compose.AsyncImage
import com.example.kalasetu.theme.KalasetuTheme

data class Comment(
    val id: Int,
    val userName: String,
    val avatarUrl: String,
    val content: String,
    val timeAgo: String,
    val likes: Int,
    val isLiked: Boolean = false,
    val repliesCount: Int = 0
)

val dummyComments = listOf(
    Comment(1, "John Doe", "https://i.pravatar.cc/150?img=12", "Lorem ipsum dolor lorem ipsum akjbd kajkjas jbkhb", "2h", 231, true, 23),
    Comment(2, "John Doe", "https://i.pravatar.cc/150?img=12", "Lorem ipsum dolor lorem ipsum akjbd kajkjas jbkhb", "2h", 231, false, 23),
    Comment(3, "John Doe", "https://i.pravatar.cc/150?img=12", "Lorem ipsum dolor lorem ipsum akjbd kajkjas jbkhb", "2h", 231, true, 23),
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CommentsBottomSheet(
    postId: Int,
    commentsList: List<Comment>,
    onSendComment: (String) -> Unit,
    onDismissRequest: () -> Unit,
    sheetState: SheetState
) {
    var mainComment by remember { mutableStateOf("") }

    ModalBottomSheet(
        onDismissRequest = onDismissRequest,
        sheetState = sheetState,
        dragHandle = { BottomSheetDefaults.DragHandle() },
        containerColor = Color.White,
        shape = RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .fillMaxHeight(0.85f)
        ) {
            // Header
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "${commentsList.size}",
                    color = Color.Gray,
                    fontSize = 14.sp,
                    modifier = Modifier.width(48.dp)
                )
                Text(
                    text = "Comments",
                    fontWeight = FontWeight.Bold,
                    fontSize = 18.sp
                )
                IconButton(
                    onClick = onDismissRequest,
                    modifier = Modifier.width(48.dp)
                ) {
                    Icon(Icons.Default.Close, contentDescription = "Close", tint = Color.LightGray)
                }
            }

            HorizontalDivider(color = Color(0xFFEEEEEE), thickness = 1.dp)

            // Comments List
            if (commentsList.isEmpty()) {
                Box(modifier = Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.Center) {
                    Text("No comments yet. Be the first to comment!", color = Color.Gray, fontSize = 14.sp)
                }
            } else {
                LazyColumn(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth(),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(24.dp)
                ) {
                    items(commentsList) { comment ->
                        CommentItem(comment)
                    }
                }
            }

            // Bottom Input
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                TextField(
                    value = mainComment,
                    onValueChange = { mainComment = it },
                    modifier = Modifier
                        .weight(1f)
                        .height(52.dp),
                    placeholder = {
                        Text(
                            text = "Add a comment...",
                            color = Color.Gray,
                            fontSize = 14.sp
                        )
                    },
                    shape = RoundedCornerShape(26.dp),
                    colors = TextFieldDefaults.colors(
                        focusedContainerColor = Color(0xFFF5F5F5),
                        unfocusedContainerColor = Color(0xFFF5F5F5),
                        disabledContainerColor = Color(0xFFF5F5F5),
                        focusedIndicatorColor = Color.Transparent,
                        unfocusedIndicatorColor = Color.Transparent,
                        disabledIndicatorColor = Color.Transparent,
                        cursorColor = Color.Black
                    ),
                    singleLine = true
                )
                Spacer(Modifier.width(8.dp))
                IconButton(
                    onClick = {
                        if (mainComment.isNotBlank()) {
                            onSendComment(mainComment.trim())
                            mainComment = ""
                        }
                    },
                    modifier = Modifier.size(44.dp).clip(CircleShape).background(Color(0xFF7466F1))
                ) {
                    Icon(Icons.AutoMirrored.Filled.Send, contentDescription = "Send", tint = Color.White, modifier = Modifier.size(20.dp))
                }
            }
            Spacer(modifier = Modifier.height(8.dp))
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Preview(showBackground = true)
@Composable
private fun CommentsBottomSheetPreview() {
    val sheetState = rememberModalBottomSheetState()
    KalasetuTheme {
        CommentsBottomSheet(
            postId = 1,
            commentsList = dummyComments,
            onSendComment = {},
            onDismissRequest = {},
            sheetState = sheetState
        )
    }
}

@Composable
fun CommentItem(comment: Comment, isReply: Boolean = false) {
    var showReplies by remember { mutableStateOf(false) }
    var isLiked by remember { mutableStateOf(comment.isLiked) }
    var likesCount by remember { mutableIntStateOf(comment.likes) }
    
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = if (isReply) 48.dp else 0.dp)
    ) {
        AsyncImage(
            model = comment.avatarUrl,
            contentDescription = null,
            modifier = Modifier
                .size(if (isReply) 28.dp else 36.dp)
                .clip(CircleShape),
            contentScale = ContentScale.Crop
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    text = comment.userName,
                    fontWeight = FontWeight.Bold,
                    fontSize = if (isReply) 12.sp else 14.sp
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = comment.timeAgo,
                    color = Color.LightGray,
                    fontSize = if (isReply) 10.sp else 12.sp
                )
            }
            Text(
                text = comment.content,
                fontSize = if (isReply) 12.sp else 14.sp,
                modifier = Modifier.padding(vertical = 4.dp),
                lineHeight = if (isReply) 16.sp else 18.sp
            )

            if (comment.repliesCount > 0 && !isReply) {
                Text(
                    text = if (showReplies) "hide replies" else "${comment.repliesCount} replies >",
                    fontWeight = FontWeight.Bold,
                    fontSize = 13.sp,
                    modifier = Modifier.padding(top = 12.dp).clickable { showReplies = !showReplies }
                )
            }
            
            if (showReplies && !isReply) {
                Column(modifier = Modifier.padding(top = 16.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
                    // Show some dummy replies
                    CommentItem(
                        comment = comment.copy(
                            userName = "Praharsha",
                            content = "This is a reply to the comment.",
                            repliesCount = 0,
                            likes = 12,
                            isLiked = false
                        ),
                        isReply = true
                    )

                    // Reply Input
                    var replyText by remember { mutableStateOf("") }
                    BasicTextField(
                        value = replyText,
                        onValueChange = { replyText = it },
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(start = 48.dp, top = 8.dp),
                        textStyle = TextStyle(
                            fontSize = 14.sp,
                            color = Color.Black
                        ),
                        cursorBrush = SolidColor(Color.Black),
                        singleLine = true,
                        decorationBox = { innerTextField ->
                            Surface(
                                color = Color(0xFFF5F5F5),
                                shape = RoundedCornerShape(22.dp),
                                modifier = Modifier.height(40.dp)
                            ) {
                                Box(
                                    modifier = Modifier.padding(horizontal = 16.dp),
                                    contentAlignment = Alignment.CenterStart
                                ) {
                                    if (replyText.isEmpty()) {
                                        Text(
                                            text = "Reply to ${comment.userName}...",
                                            color = Color.Gray,
                                            fontSize = 12.sp
                                        )
                                    }
                                    innerTextField()
                                }
                            }
                        }
                    )
                }
            }
        }
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            IconButton(
                onClick = { 
                    isLiked = !isLiked
                    if (isLiked) likesCount++ else likesCount--
                }, 
                modifier = Modifier.size(24.dp)
            ) {
                Icon(
                    imageVector = if (isLiked) Icons.Filled.Favorite else Icons.Outlined.FavoriteBorder,
                    contentDescription = "Like",
                    tint = if (isLiked) Color.Red else Color.LightGray,
                    modifier = Modifier.size(if (isReply) 16.dp else 20.dp)
                )
            }
            Text(
                text = likesCount.toString(),
                color = Color.Gray,
                fontSize = if (isReply) 9.sp else 11.sp
            )
        }
    }
}
