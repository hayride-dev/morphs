wit_bindgen::generate!({
    generate_all,
    world: "gptoss",
});

use openai_harmony::chat::{Message, Role, Conversation};
use openai_harmony::{HarmonyEncodingName, load_harmony_encoding};
use crate::hayride::ai::types::{Message as HayrideMesage, Role as HayrideRole, MessageContent};
use crate::hayride::mcp::types::CallToolParams;
use crate::exports::hayride::ai::model::{Guest, GuestFormat, GuestError,ErrorCode,Error};


struct Component;
struct Gptoss;
struct GptossError {
    message: _rt::String,
}

#[derive(Debug)]
struct ChannelSection {
    channel: String,
    content: String,
}


impl Guest for Component {
    type Format = Gptoss;
    type Error = GptossError;
}

impl GuestFormat for Gptoss {
    fn new() -> Self {
        Self
    }

    fn encode(&self,_messages:_rt::Vec::<exports::hayride::ai::model::Message>,) -> Result<_rt::Vec::<u8>,Error> {
        
        let enc = load_harmony_encoding(HarmonyEncodingName::HarmonyGptOss)
            .map_err(|_| Error::new(GptossError::new("Failed to load harmony encoding".to_string())))?;

        let mut harmony_msg = vec![];
        _messages.iter().for_each(|m| {
            m.content.iter().for_each(|c| {
                match c {
                    MessageContent::Text(t) => {
                        harmony_msg.push(Message::from_role_and_content(
                            match m.role {
                                HayrideRole::User => Role::User,
                                HayrideRole::Assistant => Role::Assistant,
                                HayrideRole::System => Role::System,
                                HayrideRole::Tool => Role::Tool,
                                HayrideRole::Unknown => Role::User,
                            },
                            t,
                        ));
                    }
                    _ => {}
                }
            });
        });

        let convo =
            Conversation::from_messages(harmony_msg);
        
        let tokens = enc.render_conversation_for_completion(&convo, Role::Assistant, None).unwrap();
        let formatted_string = enc.tokenizer().decode_utf8(&tokens).unwrap();
    
        return Ok(formatted_string.as_bytes().to_vec());
    }

    fn decode(&self,_raw:_rt::Vec::<u8>,) -> Result<exports::hayride::ai::model::Message,Error> {
        // Convert raw bytes back to string
        let formatted_string = String::from_utf8(_raw).map_err(|_| {
            Error::new(GptossError::new("Invalid UTF-8 in raw data".to_string()))
        })?;
        
        // Parse the Harmony format content
        return self.parse_harmony_content(&formatted_string);
    }
}

impl Gptoss {
    fn parse_harmony_content(&self, content: &str) -> Result<exports::hayride::ai::model::Message, Error> {
        // First try using the Harmony library's built-in parser
        let enc = load_harmony_encoding(HarmonyEncodingName::HarmonyGptOss)
            .map_err(|_| Error::new(GptossError::new("Failed to load harmony encoding".to_string())))?;
        
        // Use Harmony's tokenizer to tokenize the formatted string
        let tokens = enc.tokenizer().encode_with_special_tokens(content);
        
        // Use Harmony's proper parsing function to extract messages from completion tokens
        let parsed_messages = enc.parse_messages_from_completion_tokens(tokens, Some(Role::Assistant)).unwrap_or_else(|_| {
            // Fallback to manual parsing if Harmony fails
            Vec::new()
        });

        // Process parsed messages to find the most relevant content
        if !parsed_messages.is_empty() {
            return self.process_harmony_messages(&parsed_messages);
        }
        
        // Manual parsing fallback for streaming/partial content
        return self.parse_manual_harmony_format(content);
    }
    
    fn process_harmony_messages(&self, messages: &[Message]) -> Result<exports::hayride::ai::model::Message, Error> {
        // Track different types of content across channels
        let mut final_content = String::new();
        let mut analysis_content = String::new();
        let mut commentary_content = String::new();
        let mut tool_calls = Vec::new();
        let mut has_final_channel = false;
        
        for msg in messages {
            let channel = msg.channel.as_deref().unwrap_or("default");
            
            // Extract text content from the message
            let text_content = msg.content
                .iter()
                .filter_map(|content| {
                    match content {
                        openai_harmony::chat::Content::Text(text_content) => Some(text_content.text.clone()),
                        _ => None,
                    }
                })
                .collect::<Vec<_>>()
                .join(" ");
            
            // Check for tool calls (recipient indicates tool call)
            if let Some(recipient) = &msg.recipient {
                if recipient.starts_with("functions.") {
                    // This is a tool call
                    let tool_name = recipient.strip_prefix("functions.").unwrap_or(recipient);
                    tool_calls.push((tool_name.to_string(), text_content.clone()));
                    continue;
                }
            }
            
            // Categorize by channel
            match channel {
                "final" => {
                    final_content.push_str(&text_content);
                    has_final_channel = true;
                }
                "analysis" => {
                    analysis_content.push_str(&text_content);
                }
                "commentary" => {
                    commentary_content.push_str(&text_content);
                }
                _ => {
                    // Default channel content goes to final
                    final_content.push_str(&text_content);
                }
            }
        }
        
        // If we have tool calls, create a tool call message
        if !tool_calls.is_empty() {
            let (tool_name, tool_args) = &tool_calls[0]; // Use first tool call
            
            // Parse tool arguments from JSON string to key-value pairs
            let parsed_args = self.parse_tool_arguments(tool_args);
            
            return Ok(HayrideMesage{
                role: HayrideRole::Assistant,
                content: vec![MessageContent::ToolInput(CallToolParams {
                    name: tool_name.clone(),
                    arguments: parsed_args,
                })],
                final_: false,
            }.into());
        }
        
        // Debug: Log what we parsed (this helps with debugging)
        eprintln!("DEBUG - Parsed channels:");
        eprintln!("  Final: '{}'", final_content.trim());
        eprintln!("  Analysis: '{}'", analysis_content.trim());
        eprintln!("  Commentary: '{}'", commentary_content.trim());
        eprintln!("  Has Final Channel: {}", has_final_channel);
        
        // Build content list with multiple channels if they exist
        let mut content_items = Vec::new();
        
        // Add analysis content first if it exists (chain of thought)
        if !analysis_content.trim().is_empty() {
            content_items.push(MessageContent::Text(analysis_content.trim().to_string()));
        }
        
        // Add commentary content if it exists and is different from analysis
        if !commentary_content.trim().is_empty() && commentary_content.trim() != analysis_content.trim() {
            content_items.push(MessageContent::Text(commentary_content.trim().to_string()));
        }
        
        // Add final content (this is the main response)
        if !final_content.trim().is_empty() {
            content_items.push(MessageContent::Text(final_content.trim().to_string()));
        } else if content_items.is_empty() {
            // Fallback: if no content at all, add a placeholder
            content_items.push(MessageContent::Text("".to_string()));
        }
        
        eprintln!("  Returning {} content items", content_items.len());
        
        return Ok(HayrideMesage{
            role: HayrideRole::Assistant,
            content: content_items,
            final_: has_final_channel,
        }.into());
    }
    
    fn parse_manual_harmony_format(&self, content: &str) -> Result<exports::hayride::ai::model::Message, Error> {
        // Handle partial streaming content that might not be complete Harmony format
        
        // Look for channel markers
        if content.contains("<|channel|>") {
            return self.parse_channel_sections(content);
        }
        
        // Look for message structure markers
        if content.contains("<|start|>") || content.contains("<|message|>") {
            return self.parse_message_structure(content);
        }
        
        // If no Harmony markers, treat as plain text (streaming case)
        return Ok(HayrideMesage{
            role: HayrideRole::Assistant,
            content: vec![MessageContent::Text(content.to_string())],
            final_: false,
        }.into());
    }
    
    fn parse_channel_sections(&self, content: &str) -> Result<exports::hayride::ai::model::Message, Error> {
        let mut final_content = String::new();
        let mut analysis_content = String::new();
        let mut commentary_content = String::new();
        let mut has_final = false;
        let mut tool_calls = Vec::new();
        
        // Split by channel markers
        let parts: Vec<&str> = content.split("<|channel|>").collect();
        
        for part in parts.iter().skip(1) { // Skip the first empty part
            if part.is_empty() {
                continue;
            }
            
            // Find the channel name (first word/line)
            let lines: Vec<&str> = part.lines().collect();
            if lines.is_empty() {
                continue;
            }
            
            let channel_line = lines[0].trim();
            let channel_content = if lines.len() > 1 {
                lines[1..].join("\n").trim().to_string()
            } else {
                // Check if there's content after space/newline in the same line
                if let Some(space_pos) = channel_line.find(' ') {
                    let channel_name = &channel_line[..space_pos];
                    let content = &channel_line[space_pos + 1..];
                    match channel_name {
                        "final" => {
                            final_content = content.trim().to_string();
                            has_final = true;
                        }
                        "analysis" => {
                            analysis_content = content.trim().to_string();
                        }
                        "commentary" => {
                            commentary_content = content.trim().to_string();
                        }
                        _ => {}
                    }
                }
                continue;
            };
            
            // Check for tool call indicators in commentary channel
            if channel_line.contains("commentary") && channel_content.contains("to=functions.") {
                // Extract tool name and arguments
                if let Some(start) = channel_content.find("to=functions.") {
                    let after_to = &channel_content[start + "to=functions.".len()..];
                    if let Some(space_pos) = after_to.find(' ') {
                        let tool_name = &after_to[..space_pos];
                        let remaining = &after_to[space_pos..];
                        
                        // Look for JSON content after <|constrain|>json or <|message|>
                        let tool_args = if let Some(json_start) = remaining.find('{') {
                            if let Some(json_end) = remaining.rfind('}') {
                                &remaining[json_start..=json_end]
                            } else {
                                remaining.trim()
                            }
                        } else {
                            remaining.trim()
                        };
                        
                        tool_calls.push((tool_name.to_string(), tool_args.to_string()));
                    }
                }
            } else {
                // Regular channel content
                match channel_line.trim() {
                    "final" => {
                        final_content = channel_content;
                        has_final = true;
                    }
                    "analysis" => {
                        analysis_content = channel_content;
                    }
                    "commentary" => {
                        commentary_content = channel_content;
                    }
                    _ => {
                        // Unknown channel, add to final as fallback
                        final_content.push_str(&channel_content);
                    }
                }
            }
        }
        
        // Return tool call if found
        if !tool_calls.is_empty() {
            let (tool_name, tool_args) = &tool_calls[0];
            let parsed_args = self.parse_tool_arguments(tool_args);
            
            return Ok(HayrideMesage{
                role: HayrideRole::Assistant,
                content: vec![MessageContent::ToolInput(CallToolParams {
                    name: tool_name.clone(),
                    arguments: parsed_args,
                })],
                final_: false,
            }.into());
        }
        
        // Return the most appropriate content
        let result_content = if has_final && !final_content.trim().is_empty() {
            final_content.trim().to_string()
        } else if !commentary_content.trim().is_empty() && commentary_content.trim() != analysis_content.trim() {
            commentary_content.trim().to_string()
        } else if !final_content.trim().is_empty() {
            final_content.trim().to_string()
        } else {
            analysis_content.trim().to_string()
        };
        
        return Ok(HayrideMesage{
            role: HayrideRole::Assistant,
            content: vec![MessageContent::Text(result_content)],
            final_: has_final,
        }.into());
    }
    
    fn parse_message_structure(&self, content: &str) -> Result<exports::hayride::ai::model::Message, Error> {
        // Handle content with <|start|>, <|message|>, <|end|> structure
        
        // Extract content between <|message|> and <|end|> or <|call|> or <|return|>
        if let Some(message_start) = content.find("<|message|>") {
            let content_start = message_start + "<|message|>".len();
            let content_section = &content[content_start..];
            
            // Find the end marker
            let content_end = if let Some(end) = content_section.find("<|end|>") {
                end
            } else if let Some(call) = content_section.find("<|call|>") {
                call
            } else if let Some(ret) = content_section.find("<|return|>") {
                ret
            } else {
                content_section.len()
            };
            
            let message_content = &content_section[..content_end].trim();
            
            // Check if this is a tool call (indicated by <|call|> or "to=" in header)
            let is_tool_call = content.contains("<|call|>") || content.contains("to=functions.");
            
            if is_tool_call && content.contains("to=functions.") {
                // Extract tool name from header
                if let Some(to_start) = content.find("to=functions.") {
                    let after_to = &content[to_start + "to=functions.".len()..];
                    let tool_name = if let Some(space) = after_to.find(' ') {
                        &after_to[..space]
                    } else if let Some(pipe) = after_to.find('<') {
                        &after_to[..pipe]
                    } else {
                        after_to.trim()
                    };
                    let parsed_args = self.parse_tool_arguments(&message_content);
                    
                    return Ok(HayrideMesage{
                        role: HayrideRole::Assistant,
                        content: vec![MessageContent::ToolInput(CallToolParams {
                            name: tool_name.to_string(),
                            arguments: parsed_args,
                        })],
                        final_: false,
                    }.into());
                }
            }
            
            // Check for final channel in header
            let is_final = content.contains("final") || content.contains("<|return|>");
            
            return Ok(HayrideMesage{
                role: HayrideRole::Assistant,
                content: vec![MessageContent::Text(message_content.to_string())],
                final_: is_final,
            }.into());
        }
        
        // Fallback: treat entire content as text
        return Ok(HayrideMesage{
            role: HayrideRole::Assistant,
            content: vec![MessageContent::Text(content.trim().to_string())],
            final_: false,
        }.into());
    }

    fn parse_channel_message(&self, content: &str) -> Result<exports::hayride::ai::model::Message, Error> {
        // Legacy method - delegate to new parsing
        self.parse_channel_sections(content)
    }
    
    fn parse_tool_arguments(&self, args_json: &str) -> Vec<(String, String)> {
        // Try to parse JSON arguments into key-value pairs
        let trimmed = args_json.trim();
        
        // Handle empty or invalid JSON
        if trimmed.is_empty() || !trimmed.starts_with('{') {
            return vec![("raw".to_string(), args_json.to_string())];
        }
        
        // Simple JSON parsing for basic object structures
        // This is a basic implementation - for production, you'd want a proper JSON parser
        let mut result = Vec::new();
        
        // Remove braces and split by commas
        let content = trimmed.trim_start_matches('{').trim_end_matches('}');
        if content.trim().is_empty() {
            return result;
        }
        
        // Split by commas (this is simplified and may not handle nested objects correctly)
        for pair in content.split(',') {
            if let Some(colon_pos) = pair.find(':') {
                let key = pair[..colon_pos].trim().trim_matches('"').trim_matches('\'');
                let value = pair[colon_pos + 1..].trim().trim_matches('"').trim_matches('\'');
                
                if !key.is_empty() {
                    result.push((key.to_string(), value.to_string()));
                }
            }
        }
        
        // If parsing failed, return the raw JSON as a single argument
        if result.is_empty() {
            result.push(("json".to_string(), args_json.to_string()));
        }
        
        result
    }
    
    fn extract_channel_sections(&self, content: &str) -> Vec<ChannelSection> {
        let mut sections = Vec::new();
        
        // Find all channel tag positions
        let mut current_pos = 0;
        while let Some(start) = content[current_pos..].find("<|channel|>") {
            let absolute_start = current_pos + start;
            let tag_end = absolute_start + "<|channel|>".len();
            
            // Find the channel name (everything until the next channel tag or end of content)
            let remaining = &content[tag_end..];
            
            // Find the end of this channel section
            let content_start = tag_end;
            let content_end = if let Some(next_channel) = remaining.find("<|channel|>") {
                tag_end + next_channel
            } else {
                content.len()
            };
            
            // Extract channel content
            let section_content = &content[content_start..content_end];
            
            // Try to determine channel name and content
            // Format is typically: <|channel|>channel_name followed by content
            if let Some(first_line_end) = section_content.find('\n') {
                let channel_name = section_content[..first_line_end].trim();
                let channel_content = section_content[first_line_end..].trim();
                
                if !channel_name.is_empty() && !channel_content.is_empty() {
                    sections.push(ChannelSection {
                        channel: channel_name.to_string(),
                        content: channel_content.to_string(),
                    });
                }
            } else {
                // No newline found, treat everything as content for "default" channel
                let trimmed = section_content.trim();
                if !trimmed.is_empty() {
                    sections.push(ChannelSection {
                        channel: "default".to_string(),
                        content: trimmed.to_string(),
                    });
                }
            }
            
            current_pos = content_end;
        }
        
        sections
    }
}

impl GptossError { 
    fn new(_message:_rt::String,) -> Self {
        Self { message: _message }
    }
}

impl Into<Error> for GptossError {
    fn into(self,) ->Error {
        return Error::new(self);
    }
}

impl GuestError for GptossError {


    fn code(&self,) -> ErrorCode {
        return ErrorCode::Unknown;
    }

    fn data(&self,) -> _rt::String {
        "".to_string()
    }
}


export!(Component);