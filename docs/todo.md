# **Design Document: Enhanced `file4you` CLI Tool**

### **Project Overview**

The `file4you` CLI tool is designed to organize files and folders in a specified directory based on a user-provided configuration map that maps file extensions to folder names. This enhanced version includes LLM (Large Language Model) integration for generating and reviewing configuration maps and a suite of additional user-focused features for a more interactive, flexible, and powerful file management experience.

---

## **1. Objectives and Goals**

- **Primary Objective**: Automate and enhance file organization using a customizable, flexible CLI tool that leverages LLMs to suggest optimal organization structures.
- **Goals**:
  - Support multiple LLM models (local or API-based) via a `--model` flag.
  - Allow dynamic, interactive diff reviews and selective approval.
  - Provide a range of customization, scheduling, backup, and user feedback options.

---

## **2. Requirements**

### **2.1 Functional Requirements**

- **Config Map Generation**: Generate an ideal configuration map using an LLM based on directory contents and user preferences.
- **File Organization**: Organize files and folders based on a user-approved configuration map.
- **Diff Review and Approval**: Generate a diff between the current and proposed directory structures, allow selective approval, and iterate until the user confirms the structure.
- **LLM Model Selection**: Allow users to specify an LLM model via the `--model` flag and support various configurations (API-based or local).
- **Template-based Prompts**: Use provider-specific prompt templates written in Go template language.

### **2.2 Non-functional Requirements**

- **Portability**: Support multiple platforms (Linux, Windows, macOS).
- **Modularity**: Easily extendable to add new LLM models, configuration options, and features.
- **User-Friendly UX**: Provide a clear and interactive CLI experience with informative prompts, logs, and error messages.

---

## **3. Architecture**

### **3.1 Core Architecture**

The `file4you` tool will follow a **Ports and Adapters (Hexagonal Architecture)** pattern, with an `LLMService` interface as the primary port and provider-specific adapters for each LLM model.

### **3.2 High-Level Components**

- **CLI Interface**: Handles command-line arguments, flags, and prompts. Manages user interaction for features like diff review, config generation, and file organization.
- **LLMService Port**: An interface for LLM operations, including generating prompts based on templates and fetching responses.
- **Adapters**:
  - Each adapter implements the `LLMService` interface and handles provider-specific API calls or local model interactions.
  - A `TemplateEngine` uses Go templates for prompt generation, tailored to each LLM provider.
- **Diff Engine**: Generates a visual diff of the current vs. proposed file structure, supporting selective approval and interactive modification.
- **Backup & Rollback Module**: Backs up directory structure before changes, allowing full or partial rollbacks.

---

## **4. Detailed Feature Descriptions**

### **4.1 Core Features**

#### 4.1.1 Config Map Generation via LLM

- **Description**: Use an LLM to generate an initial configuration map based on the contents of the specified directory. The tool prompts the LLM with user-defined data (e.g., file types and desired folder structure).
- **Implementation**: Each LLM adapter uses a provider-specific template to create prompts and fetches responses for generating a map.
- **Interaction**: The user initiates map generation with a CLI command (`--generate-map`).

#### 4.1.2 Diff Review and Approval

- **Description**: Display a diff between the current file structure and the proposed structure, with options for selective approval of changes.
- **Implementation**: A diff engine compares the file system's actual state with the proposed state, highlighting changes. Users can approve/reject parts of the diff interactively.
- **Interaction**: The user enters an interactive mode for review (`--review-diff`).

#### 4.1.3 LLM Model Selection and Template-based Prompts

- **Description**: Use a `--model` flag to allow users to select an LLM provider (e.g., OpenAI, Cohere, Local Model) and load provider-specific templates.
- **Implementation**: A `NewLLMService` function instantiates the correct adapter based on the `--model` value and loads the appropriate Go template.
- **Interaction**: Specified with the `--model` flag; users can configure models and API keys in a config file.

### **4.2 Nice-to-Have Features**

#### 4.2.1 Visual Diff Tree and Selective Approval

- **Description**: Display the diff as a color-coded tree for improved clarity. Users can approve/reject specific files or folders within the tree structure.
- **Implementation**: Extend the diff engine to organize changes in a tree format, applying color codes and allowing selective approval.
- **Interaction**: Controlled within the diff review interactive mode.

#### 4.2.2 Interactive Mode with Undo/Redo

- **Description**: Step-by-step guided organization process with an option to undo/redo changes within the session.
- **Implementation**: Track changes made during the session and provide undo/redo actions.
- **Interaction**: Entered via `--interactive` mode.

#### 4.2.3 Scheduled Cleanup

- **Description**: Allow users to schedule cleanups that run in the background at specified intervals.
- **Implementation**: Integrate with system scheduling (e.g., cron for Linux/MacOS) to run `file4you` at set intervals.
- **Interaction**: Configured with a scheduling flag and interval (e.g., `--schedule daily`).

#### 4.2.4 Detailed Logging and Usage Statistics

- **Description**: Log details of each operation and track statistics (e.g., files organized, space saved).
- **Implementation**: Implement structured logging with log levels (info, warning, error) and save statistics to a report file.
- **Interaction**: Controlled via `--log-level` and `--stats` flags.

#### 4.2.5 Snapshot Backup and Rollback

- **Description**: Automatically back up the original directory structure and configuration before applying changes, allowing users to revert.
- **Implementation**: Use a backup module to snapshot directory state. Store rollback points to revert specific operations.
- **Interaction**: Run via `--backup` and `--rollback` flags.

---

## **5. Implementation Plan**

### **5.1 Milestones**

1. **Basic LLM Integration**:
   - Develop `LLMService` interface and adapter implementations.
   - Implement Go template loading for each LLM provider.
   - Integrate CLI commands for model selection and config generation.
  
2. **Diff Review and Approval**:
   - Build diff engine and interactive CLI review.
   - Support selective approval and tree-based display.

3. **Nice-to-Have Features**:
   - Implement visual diff tree, scheduled cleanup, logging, and snapshot backup.

### **5.2 Dependencies**

- **External Libraries**:
  - Go's `text/template` for prompt templates.
  - Go's CLI libraries (`flag` for basic CLI parsing or `cobra` for a more complex CLI).
- **API Dependencies**:
  - Integrate third-party API clients (e.g., OpenAI, Cohere).

### **5.3 Risks and Mitigation**

- **LLM Model Compatibility**: Different models might return responses in varying formats. To mitigate, add a response-normalizing layer.
- **Diff Complexity**: Large directories could result in complex diffs. Mitigate by supporting filtering (e.g., by file type) and partial approval.

---

## **6. Testing Strategy**

- **Unit Testing**: Test individual components (e.g., `LLMService`, template loading, diff engine).
- **Integration Testing**: Verify each provider adapter's API integration and test CLI interaction with each feature.
- **End-to-End Testing**: Simulate user flows, including diff review, selective approval, and organization, to verify complete functionality.
- **Performance Testing**: Ensure the tool handles large directories without excessive delay or memory usage.

---

## **7. User Documentation**

- **CLI Help Command**: Provide a `--help` flag explaining each command and option.
- **Configuration Guide**: Document config file structure, `--model` usage, and template customization.
- **Feature Walkthroughs**: Step-by-step guides on core and nice-to-have features.

---

## **8. Future Considerations**

- **Additional LLM Providers**: Add adapters for emerging LLM providers.
- **Plugin System**: Allow users to create and add custom adapters as plugins.
- **Enhanced Metadata Handling**: Support metadata-based organization (e.g., EXIF data for photos).
