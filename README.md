# **Large File Upload Architecture — Problem and Proposed Solutions**

## **Problem Statement**

During customer onboarding, particularly for **Know Your Customer (KYC)** verification, users are required to upload identity documents such as PAN cards, Aadhaar cards, passports, and address proofs.

These documents can be large, especially scanned PDFs and high-resolution images.

Traditional file-upload mechanisms often require users to transfer an entire file in a single HTTP request through the application backend.

This introduces several challenges:

- **Upload failures:** Unstable networks can interrupt transfers, forcing users to restart uploads.
- **Request timeouts:** Large uploads may exceed server, proxy, or gateway timeout limits.
- **Backend bottlenecks:** Routing file contents through application servers increases bandwidth, memory, and infrastructure usage.
- **Poor user experience:** Long upload times, repeated attempts, and lack of progress visibility increase frustration.
- **Customer drop-offs:** Upload failures during onboarding can lead to users abandoning the process.
- **Scalability challenges:** Supporting thousands of concurrent uploads places significant pressure on backend infrastructure.

### **Objective**

Design a **resilient, scalable, secure, and performant file-upload architecture** that supports large documents, handles network interruptions, minimizes backend resource consumption, and improves onboarding completion rates.

## **Proposed Solutions**

### **1. Chunked / Multipart Uploads**

**Problem:** Uploading an entire file in a single request means a network failure can require restarting the upload from the beginning.

**Proposed solution:**

Split large files into smaller chunks (for example, 5–10 MB) and upload each chunk independently.

- Upload chunks individually or concurrently.
- Retry only failed chunks instead of restarting the entire file.
- Track successfully uploaded chunks.
- Assemble the chunks into the final document after all parts are uploaded.
- Use cloud-native multipart upload mechanisms such as Azure Blob Storage's `Put Block` and `Put Block List`.

**Benefit:** Improved reliability, reduced retransmission, and better performance on unstable networks.

### **2. Pre-Signed URLs / Direct-to-Storage Uploads**

**Problem:** Sending large files through application servers consumes backend bandwidth, connections, memory, and processing capacity.

**Proposed solution:**

Allow clients to upload files directly to cloud object storage using temporary, restricted upload credentials.

- The client requests upload authorization from the backend.
- The backend generates a short-lived pre-signed URL or Azure SAS URL.
- The client uploads the document directly to object storage.
- The backend remains responsible for authorization and upload verification.
- Supported storage platforms include AWS S3, Azure Blob Storage, and Google Cloud Storage.

**Benefit:** Reduces backend load, eliminates unnecessary network hops, and improves scalability.

### **3. Asynchronous File Processing**

**Problem:** Document validation, virus scanning, OCR, and KYC verification can be time-consuming. Performing these operations synchronously delays responses and increases the likelihood of timeouts.

**Proposed solution:**

Separate document uploading from document processing using an asynchronous, event-driven architecture.

- Store the uploaded document in object storage.
- Publish an event or message after successful upload verification.
- Use a message broker such as Azure Queue Storage, AWS SQS, Kafka, or RabbitMQ.
- Process documents using background workers.
- Perform malware scanning, content validation, OCR, and KYC verification.
- Track processing status and notify the frontend when processing finishes.

**Benefit:** Improves responsiveness, supports independent scaling, and isolates upload traffic from processing workloads.

### **4. Client-Side Compression and File Optimization**

**Problem:** High-resolution images and unnecessarily large documents consume additional bandwidth and increase upload times.

**Proposed solution:**

Optimize supported files on the client before uploading.

- Compress JPEG and PNG images where appropriate.
- Resize images that exceed the required resolution.
- Remove unnecessary image metadata where permitted.
- Use libraries such as `compressor.js` for browser-based image optimization.
- Preserve sufficient image quality for OCR and identity verification.

**Benefit:** Reduces transferred data, upload duration, and bandwidth consumption.

**Consideration:** Compression must not remove important identity-document details or compromise verification accuracy.

### **5. File Size Limits and Validation**

**Problem:** Oversized, unsupported, corrupted, or malicious documents can waste infrastructure resources and introduce security risks.

**Proposed solution:**

Enforce validation at both the client and server/storage boundaries.

- Define maximum file sizes, such as 25 MB.
- Restrict supported formats to PDF, JPEG, and PNG.
- Validate file extensions, MIME types, and actual file signatures.
- Verify the actual uploaded object size.
- Use checksums to detect transfer corruption.
- Scan uploaded documents for malware.
- Reject invalid documents before allowing downstream processing.

**Benefit:** Improves security, prevents unnecessary resource usage, and increases processing reliability.

### **6. Automatic Upload Resume**

**Problem:** Users with unstable mobile or broadband connections may lose connectivity during an upload. Restarting the entire transfer wastes time and bandwidth.

**Proposed solution:**

Implement resumable uploads using persisted upload-session state.

- Assign a unique identifier to each upload session.
- Track successfully uploaded chunks.
- Persist upload progress locally and/or in the backend.
- Detect interrupted uploads.
- Query the storage provider to identify completed chunks.
- Resume uploading only missing or failed chunks.
- Expire and clean up abandoned upload sessions.

**Benefit:** Improves reliability on unstable networks, reduces repeated transfers, and minimizes onboarding abandonment.

### **7. CDN / Edge-Assisted Upload Acceleration**

**Problem:** Users located far from the cloud storage region may experience increased latency and slower uploads.

**Proposed solution:**

Use upload acceleration services or regional ingestion endpoints to reduce the network distance between users and the upload infrastructure.

- Route uploads through geographically nearby edge endpoints where supported.
- Use cloud-provider upload acceleration capabilities, such as AWS S3 Transfer Acceleration.
- Optimize routing between edge locations and object storage.
- Consider regional upload endpoints for geographically distributed users.
- Evaluate additional costs, data residency, and regulatory requirements.

**Benefit:** Can reduce upload latency and improve throughput for geographically distributed customers.

**Consideration:** A conventional CDN primarily accelerates content delivery and does not automatically accelerate uploads. Upload-aware edge services are required.

## **Solution Summary**

| # | Solution                         | Primary problem addressed                    |
| - | -------------------------------- | -------------------------------------------- |
| 1 | Chunked / Multipart Uploads      | Failures requiring complete re-upload        |
| 2 | Pre-Signed URLs / Direct Uploads | Backend bottlenecks and scalability          |
| 3 | Asynchronous Processing          | Long-running validation and processing       |
| 4 | Client-Side Compression          | Excessive file sizes and bandwidth usage     |
| 5 | File Size Limits & Validation    | Invalid, corrupted, or malicious files       |
| 6 | Automatic Upload Resume          | Network interruptions and repeated transfers |
| 7 | CDN / Edge Upload Acceleration   | Geographic latency and slow transfers        |

## **Expected Outcome**

By combining these solutions, the system aims to achieve:

- **Reliability:** Uploads can recover from transient network failures.
- **Scalability:** Application servers are decoupled from large file transfers.
- **Performance:** Files transfer efficiently without unnecessary processing or network hops.
- **Security:** Documents are validated, protected, and scanned before use.
- **User experience:** Progress visibility, retries, and resume functionality reduce frustration.
- **Operational efficiency:** Uploads and document processing scale independently.

The ultimate goal is to **reduce upload failures and customer drop-offs while supporting secure, high-volume document onboarding**.
