# Guided Study Test 1 — Failure Log

## Test context

- Book: *AWS Certified Developer Study Guide*
- Requested start: Chapter 2, from the beginning
- Study session: `Chapter 2 — Compute and Networking`
- Initial page index: 110
- Current cursor after the events below: page index 111
- Purpose: evaluate an experimental, page-by-page study method and preserve failures for later developer handoff

## Failure event 1 — Learner-facing response was hidden in transient commentary

### What happened

After the learner answered the first question with “run code on AWS,” the tutor put the substantive response in transient commentary while advancing the page. In the rendered experience, that response was hidden/collapsed under reasoning rather than remaining visible as part of the teaching conversation.

### Why this is a failure

Teaching feedback is part of the durable learner-facing exchange. The learner should not have to expand reasoning or tool activity to see whether an answer was correct, why it was correct, or what happens next.

### Expected behavior

- Put answer evaluation, correction, explanation, and transition language in the visible final response.
- Keep commentary limited to operational updates that are not required to understand the lesson.
- Never make the final response depend on collapsed commentary.

## Event 2 — Reclassified observation: tail of the chapter introduction was not discussed

### What happened

After the learner identified the shared purpose of the listed services as running code on AWS, the tutor advanced to page 111. It did not explicitly discuss the remaining introductory material at the top of that page before asking about the difference between virtualized and bare-metal EC2 instances.

The material not discussed was:

- the final three entries in the opening compute-service list: AWS Lambda, VMware Cloud, and Amazon App Runner;
- the chapter roadmap explaining that the chapter will cover EC2 instances, instance customization, VPC network controls, and management concerns; and
- the statement that EC2 and VPC are foundational services whose concepts transfer to other AWS services.

### Current assessment

This is not currently classified as a failure. The omitted material is introductory orientation rather than a substantive mechanism or distinction. The tutor’s next question targeted the first concept on the page with stronger retrieval value: how virtualized EC2 instances differ from bare-metal instances.

Not forcing the learner to restate a service list or chapter roadmap can be sound educator behavior. A brief transition or summary might improve transparency, but the absence of one is not enough evidence by itself to establish a failure. This observation should be reconsidered only if a broader pattern shows that the method regularly skips substantive content, or if the learner explicitly requests exhaustive coverage.

## Failure event 3 — Diagnostic failures were nearly written into the study progress log

### What happened

When the learner asked to record and track method failures, the tutor announced that it would use the Guided Study session’s durable progress log for “diagnostic event 1.” It then read session progress in preparation for a log mutation.

The turn was interrupted before `log_progress` was called, so no diagnostic entry was written to the study service.

### Why this is a failure

The study progress log represents learning continuation state. Test-harness failures are a separate concern and should not contaminate learner progress. The tutor also chose a storage mechanism before confirming the intended diagnostic destination, despite the request being about experimental-method testing.

### Expected behavior

- Keep pedagogical progress and test diagnostics in separate stores.
- Record test failures in this file, not in the Guided Study session log.
- Use the study log only for what was learned and the exact continuation point.
- If no diagnostic destination has been specified, keep a visible in-chat incident list or ask where the test log belongs before mutating durable state.

## Failure event 4 — The reading question was too coarse and jumped to the end of the page

### What happened

On page 111, the tutor asked the learner to compare a standard virtualized EC2 instance with a bare-metal instance. The comparison required reading material near the end of the page and bypassed several earlier substantive EC2 statements as separate attention targets.

### Why this is a failure

The study method is intended to help the learner focus while reading and prove understanding through frequent questions. The tutor treated the question as a higher-level challenge or synthesis prompt instead of using fine-grained questions to guide attention through the text in sequence. As a result, important statements about EC2 instances could be read without the learner having to demonstrate understanding of them.

### Expected behavior

- Ask frequent, fine-grained questions that follow the substantive text in reading order.
- Give each question one small fact, relationship, or claim to retrieve.
- Resolve the current question before asking the next one.
- Use questions to focus the learner’s attention and verify comprehension, not primarily to create a difficult synthesis challenge.
- Do not jump to a later comparison while earlier substantive claims remain unchecked.

## Failure event 5 — The corrected fine-grained sequence still skipped the acronym definition

### What happened

After the learner explicitly required finer-grained questions in reading order, the tutor restarted the EC2 section by asking what EC2 calls the computing environments it provisions. It failed to first ask what “EC2” stands for, even though “Amazon Elastic Compute Cloud (EC2)” appears before the statement about instances.

### Why this is a failure

The tutor did not apply the newly clarified method at the granularity requested. It skipped the first central definitional mapping in the substantive section and moved directly to the next fact. This occurred immediately after the learner explained that questions must force close attention to the text.

### Expected behavior

- Begin with the earliest central fact in the substantive text.
- Treat the expansion of a central service acronym as a valid fine-grained comprehension checkpoint.
- Do not move to the next sentence-level fact until the preceding fact has been checked.

## Failure event 6 — Question design was not calibrated around thorough coverage

### What happened

After being told to make the questions more fine-grained, the tutor asked only short factual questions and then interpreted the learner’s concern as a prohibition on one-word answers. That interpretation was incorrect: one-word answers are acceptable when appropriate, but the overall questioning must thoroughly check the text.

### Why this is a failure

The tutor focused on answer length instead of coverage. The intended method uses frequent questions to make the learner read closely and demonstrate understanding. Some facts naturally require only one word or a short phrase; relationships may require one or two short sentences. The failure is an insufficiently thorough or poorly sequenced set of questions, not brevity by itself.

### Expected behavior

- Ask enough questions to check the important content thoroughly and in sequence.
- Expect a one-word or short-phrase answer for a simple fact when appropriate.
- Expect no more than one or two short sentences for a relationship or explanation.
- Choose answer length according to the content being checked rather than enforcing one response format.

## Failure event 7 — A question merely removed words from the source sentence

### What happened

The tutor asked, “What can you choose about the hardware of an EC2 instance?” The source sentence directly says that the learner can choose the hardware resources they need, so the question functioned as the next sentence with a piece removed. The learner correctly answered “the resources you need” but found the question unhelpful.

### Why this is a failure

The question checked whether the learner could copy or complete the sentence rather than whether they understood its meaning. Although simple factual questions are allowed, systematically converting source sentences into cloze prompts does not provide useful attention guidance or evidence of comprehension.

### Expected behavior

- Use direct recall for central names, definitions, and facts when that retrieval is itself valuable.
- For explanatory text, ask about meaning, relationships, distinctions, consequences, or a nearby application.
- Keep the source scope small and the expected answer short without merely deleting words from the sentence.
- Credit a correct answer and move forward rather than forcing the learner to repeat it in a preferred form.
