# What Should Become a Flashcard?
## Retrieval Practice, Knowledge Type, and the Boundary Between Remembering and Doing

### Abstract

Question-and-answer flashcards are a powerful mechanism for maintaining knowledge over time, but their usefulness varies dramatically depending on what a learner ultimately needs to do with that knowledge. This review asks a specific question: **after a learner already understands material from a textbook, which parts should be maintained through Q&A retrieval practice, which require exercises or other forms of performance practice, and which are better left as external reference material?**

The research strongly supports retrieval practice as a method for durable retention. Meta-analytic and classroom evidence shows that attempting to recall learned information generally produces better delayed retention than restudying it, with effortful recall particularly effective. Retrieval practice can also support transfer and conceptual learning. It is therefore incorrect to characterize flashcards as useful only for trivial or arbitrary facts.

However, transfer is conditional. Pan and Rickard's meta-analysis of 192 transfer effects from 122 experiments found a moderate overall transfer advantage for testing over non-testing reexposure, *d* = 0.40, but transfer depended strongly on how the retrieval task related to the eventual target task. Transfer was stronger to application and inference questions and medical diagnosis, and weaker for some problem types, including worked-example problems. Response congruency, elaboration during retrieval, and successful initial retrieval were major moderators.

This leads to the central conclusion of this review:

**Flashcards are most appropriate when the capability worth preserving is substantially a capability to retrieve a mental representation. They become progressively less sufficient as successful performance requires selecting, constructing, calculating, deriving, manipulating, diagnosing, debugging, evaluating, or arguing.**

A second distinction is equally important. Not all information that is useful is worth maintaining internally. Large tables, volatile details, rarely used parameters, and information whose exact value can cheaply be retrieved from a reliable external source may be better treated as reference material. External memory use is a normal part of human cognition rather than automatically a learning failure.

The resulting picture is therefore not “flashcards versus understanding.” Understanding is assumed. The relevant distinction is between **remembering, doing, and looking up**.

---

## 1. Scope and Assumptions

This review concerns adult knowledge professionals learning technical or intellectually demanding material in mathematics, computer science and programming, biology, medicine, chemistry, physics, economics, history, politics, and law. Language learning is excluded because vocabulary and language acquisition create a substantially different problem.

Only conventional **question-and-answer retrieval cards** are considered. Cloze deletion, image occlusion, recognition-based multiple choice, and other formats fall outside the primary scope.

Most importantly, the learner is assumed already to **understand the material**. The problem is not initial instruction. It is retention after instruction.

This assumption matters. Much advice about flashcards begins with warnings not to memorize material one does not understand. That warning is sensible but does not answer the harder downstream question. A person may understand a chapter perfectly today and nevertheless forget most of its useful contents over the next several years. Which components should be deliberately kept accessible?

The target outcome here is not merely examination performance. It is **durable retention that contributes to future understanding, reasoning, and professional use**.

The research literature does not map perfectly onto this goal. Much retrieval-practice research involves students, relatively short retention intervals, quizzes rather than literal flashcard systems, and educational tests rather than professional performance. Classroom reviews nevertheless find benefits across educational levels, subjects, retrieval formats, and retention intervals, while also emphasizing considerable heterogeneity in study designs.

Accordingly, some conclusions below are strong empirical findings, whereas others are evidence-informed synthesis. Those should not be confused.

---

# 2. What Retrieval Practice Actually Does Well

Retrieval practice is one of the most replicated findings in learning science. Attempting to retrieve previously learned information strengthens later accessibility of that information more effectively than simply exposing oneself to it again. Rowland's meta-analysis found broad support for this testing effect and also found that recall tests generally produced larger benefits than recognition tests, consistent with a role for effortful retrieval.

The effect can survive meaningful delays. In a simulated classroom experiment using history lectures, Butler and Roediger compared studying lecture summaries, multiple-choice testing, and short-answer testing. One month later, initial short-answer testing produced the greatest final recall.

Nor is retrieval restricted to verbatim factual memory. Karpicke and Blunt found retrieval practice superior to concept mapping for later learning from science texts, including measures intended to assess comprehension and inference.

This matters because one tempting taxonomy would classify “facts” as flashcard material and “concepts” as non-flashcard material. The evidence does not support such a clean split.

A question such as:

**Why does increasing extracellular potassium reduce the potassium concentration gradient across a cell membrane?**

requires retrieval of a causal relationship. If the learner genuinely understands that relationship, repeatedly retrieving it can help preserve it. It is not rendered inappropriate merely because the answer is conceptual rather than a date or definition.

The better question is: **what mental operation does answering the card rehearse?**

If the future need is to retrieve that same relation, the alignment is excellent.

If the future need is to analyze an unfamiliar electrophysiological situation involving multiple simultaneous changes, the card has preserved useful ingredients, but not the whole competence.

---

# 3. The Crucial Limitation: Transfer Is Real, but Not Automatic

The most consequential finding for flashcard design is that strengthening memory for something does not automatically strengthen every activity that depends on it.

Pan and Rickard reviewed more than 40 years of testing-and-transfer research. Across 192 effect sizes, retrieval practice did produce transfer beyond the exact practiced material, with an overall effect of *d* = 0.40 compared with non-testing reexposure. But the magnitude of transfer varied greatly. Transfer was especially evident across test formats and into some application, inference, and diagnostic tasks. It was weaker in other circumstances, and the relationship between the retrieval response and target response strongly moderated the result.

Agarwal's experiments examining fact-level and higher-order retrieval make the practical consequence unusually clear. Fact-based retrieval improved subsequent factual performance. Higher-order retrieval improved higher-order performance. Merely strengthening factual retrieval did **not** automatically confer an equivalent higher-order benefit.

A similar pattern appeared in classroom science research: questions requiring application could benefit later application, whereas definition questions did not necessarily improve performance on application questions.

This gives us a more useful formulation than “retrieval practice works”:

**Retrieval practice trains retrieval, and the consequences of that training depend on what is retrieved and how similar that retrieval operation is to future use.**

A card asking “What is Bayes' theorem?” can maintain Bayes' theorem.

A card asking “Under what conditions does Bayes' theorem substantially revise a prior probability?” can maintain conditional conceptual knowledge.

Neither guarantees that a physician will correctly combine likelihoods in a messy diagnostic case or that an engineer will identify the relevant probability model in an unfamiliar reliability problem.

Those later performances contain additional operations.

---

# 4. A Knowledge-Type Account of Flashcard Appropriateness

The evidence suggests organizing textbook content by the **kind of knowledge and future performance involved**, rather than primarily by academic discipline.

| Knowledge type | Q&A flashcard suitability | What cards can preserve | What usually requires another activity |
|---|---|---|---|
| Atomic declarative knowledge | **High** | Facts, terms, constants, labels, compact propositions | Usually little, unless facts must be applied |
| Relational / causal knowledge | **Moderate to high** | Mechanisms, relationships, contrasts, explanations | Novel application, model use, multi-factor reasoning |
| Conditional / discriminative knowledge | **Moderate** | Conditions, cues, distinctions, applicability rules | Selecting the right rule in unfamiliar mixed cases |
| Frequently used foundational knowledge | **Often high** | Fast access to recurring cognitive building blocks | Application still requires practice |
| Procedures / algorithms | **Supportive only** | Steps, invariants, constraints, purposes of steps | Actually executing the procedure |
| Mathematical derivations / proofs | **Supportive only** | Theorem statements, assumptions, key identities, strategic ideas | Reconstructing derivations and producing proofs |
| Problem-solving schemas | **Conditional** | Recognition cues and strategic principles | Solving varied problems |
| Complex judgment | **Low as primary method** | Relevant criteria, factual inputs, diagnostic features | Cases, comparison, evaluation, argument and decisions |
| Large or volatile reference information | **Usually very low** | Selected landmarks or qualitative patterns | Lookup from authoritative reference |

The rest of this review develops these categories.

---

# 5. Atomic Declarative Knowledge: The Strongest Case for Cards

Atomic declarative knowledge consists of compact propositions whose future use primarily requires remembering them.

Examples include anatomical relationships, terminology, definitions, named laws, historical chronology, institutional structures, mathematical notation, stable programming concepts, chemical properties, and factual associations.

This category has the strongest direct relationship with traditional Q&A flashcards.

In anatomy and physiology, for example, retrieval practice has repeatedly improved retention of learned information. Dobson concluded that retrieval practice was an efficient and effective method for retaining anatomy and physiology material, and later work found advantages for combining distributed practice with retrieval.

A biology learner might therefore reasonably maintain:

**Q: Which organelle is the primary site of oxidative phosphorylation in eukaryotic cells?  
A: The inner mitochondrial membrane.**

A physician might maintain:

**Q: Which cranial nerve carries the major parasympathetic innervation to thoracic and abdominal viscera?  
A: The vagus nerve, CN X.**

A historian might maintain:

**Q: What institutional change did the Seventeenth Amendment make to the U.S. Senate?  
A: It established direct election of senators.**

A programmer might maintain:

**Q: What asymptotic lookup complexity does a well-behaved hash table provide on average?  
A: Expected O(1).**

These are not all “arbitrary facts.” Some participate in conceptual structures. What they share is that **future availability depends heavily on being able to call the proposition to mind**.

For this knowledge type, Q&A retrieval is unusually well aligned with the desired competence.

But even here, there is a second filter: **is this fact worth internalizing at all?**

That question becomes important later.

---

# 6. Relational and Causal Knowledge: Often Good Card Material

Technical expertise is full of relationships rather than isolated facts:

pressure varies with volume under specified conditions; increasing interest rates affects borrowing incentives; a particular mutation alters a protein pathway; a database index changes the cost profile of particular queries; a constitutional doctrine constrains a particular institutional action.

Cards can maintain these relationships surprisingly well if the question actually requires retrieval of the relationship.

For example:

**Q: Why does increasing temperature generally increase the rate constant of a chemical reaction?  
A: A larger fraction of molecular collisions have sufficient energy to overcome the activation barrier.**

That is a useful Q&A retrieval item even though the answer expresses a causal mechanism.

Similarly:

**Q: Why can increasing the number of indexes on a database slow writes?  
A: Each write may require maintaining additional index structures.**

Or:

**Q: Why does an unexpected increase in inflation tend to redistribute wealth from creditors to debtors when debts are nominally fixed?  
A: Repayment occurs in money with lower real purchasing power than anticipated.**

The science-text findings from Karpicke and Blunt, as well as broader retrieval-transfer research, make it difficult to sustain the view that cards should be restricted to meaningless facts. Retrieval can strengthen meaningful, interconnected knowledge.

The limitation is scope.

Remembering the causal proposition about inflation does not train the learner to analyze a complicated macroeconomic episode involving supply shocks, expectations, monetary policy, wage contracts, and institutional context.

A relational card can keep a **model component** alive. It does not necessarily keep the ability to operate an entire model alive.

That makes conceptual cards appropriate when the relationship itself is an important unit of thought.

---

# 7. Conditional and Discriminative Knowledge: Cards Help, but Cases Matter

Many professional failures occur not because someone has forgotten a rule, but because they fail to recognize **when the rule applies**.

This is conditional knowledge.

Examples include knowing when conservation of momentum is an appropriate simplifying principle, when integration by parts is promising, when a database transaction needs a stronger isolation guarantee, when a patient presentation suggests one diagnosis rather than another, or when a legal precedent is meaningfully analogous.

A card can directly target this:

**Q: When is conservation of linear momentum applicable to a system over an interval?  
A: When net external impulse over the interval is negligible.**

That seems substantially better than merely memorizing:

**Q: What is conservation of momentum?**

But even the improved card does not fully train **selection among competing methods**.

Mathematics research on interleaved practice is particularly informative here. When problem types are mixed rather than grouped by method, learners must discriminate among problems and select an appropriate strategy. Rohrer and colleagues found benefits of interleaving that appear to involve both distinguishing problem types and strengthening associations between problem features and solution strategies.

Consider calculus.

A card can preserve:

**Q: What structural feature often makes integration by parts useful?  
A: A product where differentiating one factor simplifies it while the other can be integrated.**

Useful.

But the actual professional competence is closer to being shown an integral with no chapter heading and deciding whether to use substitution, integration by parts, partial fractions, an identity, numerical integration, or something else.

That requires **mixed problem selection practice**.

The same principle generalizes.

In medicine, cards can preserve discriminative features of diseases, but diagnosis requires selecting among competing explanations from a case.

In programming, cards can preserve tradeoffs between data structures, while architecture exercises require choosing one under actual constraints.

In law, cards can preserve doctrinal elements, while hypothetical cases require identifying which doctrines matter.

In history, cards can preserve characteristics of different source types, but historical reasoning requires evaluating actual sources.

Conditional cards are therefore often useful, but they should not be mistaken for the performance of conditional reasoning itself.

---

# 8. Frequently Used Foundational Knowledge: When Derivable Knowledge May Still Be Worth Memorizing

A more difficult category is knowledge that **could be reconstructed**, but whose repeated reconstruction may be inefficient.

The cleanest evidence comes from arithmetic and mathematical fluency. Research on computational automaticity describes reductions in working-memory demands as elementary computations become automatic, while working memory is strongly related to arithmetic performance more generally.

This gives some scientific substance to the intuition that certain internal knowledge functions as the **working vocabulary of reasoning**.

For example, a mathematician does not need to derive from first principles every time that

\[
\frac{d}{dx}e^x=e^x.
\]

A programmer should not need documentation to remember what a stack's basic access discipline is.

A chemist gains little by repeatedly reconstructing the qualitative meaning of electronegativity.

A physician cannot plausibly pause to consult a reference for every foundational anatomical relationship involved in interpreting a case.

Instant availability can lower friction in a larger chain of reasoning.

But caution is necessary. Evidence that automatic arithmetic supports mathematical performance does **not** imply that every formula in a technical textbook should be memorized.

A reasonable interpretation is narrower:

A derivable item becomes a stronger candidate for long-term retrieval when it is **compact, stable, repeatedly needed, and functions as an input to many other cognitive operations**.

This is why remembering the quadratic formula may be defensible while memorizing the closed form of an obscure special-function identity encountered once in a textbook may be ridiculous.

The criterion is not “can this be derived?”

It is closer to:

**How expensive is it to repeatedly derive or retrieve externally, and how often does immediate availability improve subsequent thought?**

---

# 9. Procedures and Algorithms: Remembering the Recipe Is Not the Skill

Procedural knowledge is where static Q&A cards begin to run into a fundamental mismatch.

A card can ask:

**Q: What invariant does binary search maintain?  
A: If the target exists, it remains within the current candidate interval.**

Excellent.

Another can ask:

**Q: What is the first step in Gaussian elimination?  
A: Select a pivot and use row operations to eliminate entries below it.**

Also potentially useful.

But neither card trains someone to execute binary search correctly under unusual boundary conditions or to perform Gaussian elimination on a messy system.

Research comparing retrieval practice with worked examples and problem solving illustrates the difficulty. In mathematical word-problem research, retrieval after worked-example study has not always improved later problem-solving performance compared with restudy, and the literature has generated an active debate over when retrieval practice helps procedural learning.

Programming education reaches a similar conclusion from another direction. Reviews of worked examples and research on Parsons problems distinguish activities such as tracing, rearranging, completing, and independently writing code. These tasks impose increasingly authentic demands on program construction. Simply recalling statements about a loop or recursion is not equivalent to producing working code.

Medicine supplies an especially vivid example. Testing can improve retention of **skills**, but when researchers test a skill such as CPR, the “test” consists of performing CPR. A six-month CPR study found a practically meaningful but statistically inconclusive advantage for a course ending in skills testing over equivalent extra practice, with an effect size of about 0.4.

This is a crucial conceptual point:

**“Retrieval practice” does not necessarily mean answering verbal questions. For a skill, an ecologically aligned retrieval test may be performing the skill.**

If the thing worth retaining is “how to debug a race condition,” a question asking for the steps of debugging may help maintain a checklist. Debugging actual programs is still required to preserve debugging competence.

---

# 10. Mathematical Derivations and Proofs: Preserve Ingredients, Practice Reconstruction

Proofs and derivations are unusually tempting targets for bad flashcards.

One could make:

**Q: Prove the spectral theorem.  
A: [three paragraphs of proof].**

Technically, this is a question-and-answer card. Pedagogically, it is mostly a small hostage situation.

The answer contains too much structure for a conventional card to produce useful, unambiguous retrieval. More importantly, successful recognition of a memorized proof is not equivalent to being able to reconstruct the argument.

Q&A cards can instead preserve components:

**Q: What property of a real symmetric matrix guarantees orthogonality of eigenvectors associated with distinct eigenvalues?  
A: Symmetry, via the equality \(v^TAw=(Av)^Tw\).**

or:

**Q: What is the central substitution used to derive the relativistic energy-momentum relation from four-momentum?  
A: Express four-momentum using \(p^\mu = m u^\mu\) and apply the invariant Minkowski norm.**

These maintain landmarks.

But if the goal is to retain the derivation itself, the learner periodically needs to **reconstruct the derivation without the solution visible**.

Physics education research is useful here. Gjerde found that retrieving relevant physics principles and their conditions of application before self-explaining worked examples improved aspects of subsequent problem solving and the quality of self-explanation. This is not evidence that cards replace problems. It is evidence that accessible principles can support richer practice.

That is probably the right relationship.

Cards preserve the tools and strategic landmarks.

Derivations preserve the ability to assemble them.

---

# 11. Problem-Solving Schemas: A Supporting Role for Cards

Some textbook knowledge is neither a fact nor a fixed procedure. It is a reusable **problem-solving schema**.

Examples include recognizing conservation-law problems in physics, translating a word problem into a probabilistic model, decomposing a dynamic-programming problem into overlapping subproblems, or identifying an endogeneity problem in econometrics.

A retrieval card can maintain the abstract schema:

**Q: What two properties make dynamic programming a promising strategy?  
A: Overlapping subproblems and exploitable optimal substructure.**

That is valuable.

But solving an unfamiliar dynamic-programming problem additionally requires identifying states, transitions, base cases, evaluation order, and complexity. Those operations are learned and maintained by solving varied problems.

This distinction explains why generic claims such as “retrieval practice improves problem solving” are too crude.

If retrieval practice consists of **solving a problem from memory**, then problem solving itself can be retrieval practice.

If retrieval practice consists of recalling a one-sentence description of dynamic programming, its likely transfer target is much narrower.

Freeman and colleagues' meta-analysis of 225 undergraduate STEM studies found that active-learning environments substantially improved student performance and lowered failure rates relative to traditional lecturing. That literature encompasses many forms of active reasoning and problem solving rather than simple recall.

For maintaining problem-solving competence, therefore, Q&A cards are best seen as a mechanism for **keeping the relevant conceptual machinery accessible**, not replacing the use of that machinery.

---

# 12. Complex Judgment: Cases Beat Cards as the Primary Maintenance Activity

Professional judgment requires combining facts and concepts under uncertainty.

Diagnosis, legal analysis, historical interpretation, policy analysis, and engineering design all fall here.

Medicine offers direct experimental evidence. Sheldon and colleagues compared passive review of patient cases with active retrieval of case information and examined subsequent diagnostic transfer. Active retrieval from cases improved transfer to novel diagnostic decisions, illustrating how retrieval can support complex reasoning when what is retrieved resembles the eventual reasoning task.

Notice what this does **not** imply.

It does not imply that a deck of cards such as:

**Q: What disease causes symptom X?  
A: Disease Y.**

is equivalent to case-based diagnostic reasoning.

The successful retrieval activity involved cases.

Likewise, legal reasoning depends centrally on comparison, analogy, distinction, and the application of rules to facts. Work on legal analogy emphasizes that analogical reasoning depends on perceiving legally relevant similarities through domain knowledge and experience.

Cards could preserve:

**Q: What are the elements of negligence?**

or:

**Q: What level of scrutiny applies to a racial classification under U.S. equal-protection doctrine?**

But maintaining the ability to reason legally requires hypotheticals and cases in which the learner must decide what matters.

History has the same structure. The Butler and Roediger history experiment provides good evidence for retaining lecture content using short-answer retrieval over a month.

But evaluating whether two historical sources corroborate one another is a different task. Remembering who signed a treaty is retrieval. Evaluating why accounts of the treaty conflict is historical reasoning.

Cards preserve the **knowledge base from which judgment operates**. Judgment itself requires judgment-shaped practice.

---

# 13. Reference Knowledge: The Case for Deliberately Not Memorizing

A learning system can fail by forgetting too much.

It can also fail by demanding that the learner remember far too much.

Consider a table listing the density of water at dozens of temperatures. The exact entries are useful. They are also cheap to retrieve from an authoritative reference and rarely need to be available simultaneously in unaided memory.

Converting every row into a card would create review burden without proportionate cognitive value.

The useful long-term knowledge is more likely to be something like:

**Q: Around what temperature does liquid water reach its maximum density at standard pressure?  
A: Approximately 4°C.**

and perhaps:

**Q: Why is water's density-temperature relationship environmentally important?  
A: Because the density maximum near 4°C affects stratification, freezing behavior, and circulation in bodies of water.**

The table remains a table.

Research on cognitive offloading reinforces the broader point that humans strategically use external tools to reduce internal memory demands. Offloading is not inherently pathological; it is a normal extension of cognitive work, although people can make imperfect judgments about when to rely on it.

For knowledge professionals, reference-only information often includes exhaustive physical property tables, complete drug databases, full statutory language, current tax rates, API option catalogs, hardware specifications, complete historical datasets, and rarely used constants.

The case becomes even stronger when information is **volatile**.

Memorizing a stable mathematical identity creates little risk of future staleness.

Memorizing every parameter of a cloud platform's current API may leave a professional confidently remembering obsolete information three years later, arguably the least charming form of expertise.

---

# 14. What Makes Knowledge Worth Internalizing?

The evidence does not provide a validated seven-question algorithm for deciding whether a textbook sentence deserves a flashcard. The following synthesis should therefore be treated as a **descriptive framework**, not as a law of learning.

The value of internal retrieval appears greatest when several conditions converge.

Knowledge is a stronger candidate for a Q&A retrieval card when **future use itself requires recall**, when it is a frequent prerequisite for other reasoning, when having it immediately available reduces interruption or working-memory burden, when it helps recognize patterns or errors, when it is stable enough to remain correct, and when the unit of knowledge is compact enough that maintaining it has reasonable cost.

Conversely, flashcard value declines when information is rarely used, extremely detailed, volatile, inexpensive to look up, or meaningful primarily as part of a performance that a card does not exercise.

This distinction can be summarized as:

**Internalize the reusable inputs to thought. Practice the operations of thought. Externalize the archive.**

That sentence is a synthesis of the evidence, not an experimentally established rule.

---

# 15. Cross-Domain Examples

## Mathematics

A good card might ask:

**What hypotheses are required for Rolle's theorem?**

The hypotheses themselves are compact and need to be available when reasoning about functions.

Another good candidate:

**What does a negative determinant tell you about the orientation of a linear transformation?**

This preserves a conceptual relation.

By contrast, the ability to determine which theorem can solve an unfamiliar proof problem requires proof practice. The ability to solve differential equations requires solving differential equations. Mixed mathematical practice is particularly important when the challenge is selecting among strategies.

A table of 100 numerical approximations of special functions should generally remain reference material.

### Practical boundary

**Remember definitions, conditions, central identities, conceptual relations, and sufficiently frequent mathematical subroutines. Practice derivation, proof, strategy selection, and problem solving. Look up rare constants and specialized formulas.**

---

## Physics

Good cards include fundamental principles and their applicability conditions:

**When is mechanical energy conserved?**

**What physical quantity does the area under a force-position curve represent?**

Such knowledge must remain available during problem solving.

But predicting the motion of a multi-body system, choosing an approximation, constructing a free-body diagram, or deriving an expression requires performance practice. Physics work on retrieval plus self-explanation suggests that accessible principles can improve subsequent work with examples, rather than making examples unnecessary.

### Practical boundary

**Card the laws, meanings, conditions, qualitative relationships, and frequently used formulas. Solve problems to maintain model construction and application.**

---

## Chemistry

A useful card:

**Why does a catalyst change reaction rate without changing the equilibrium constant?**

This retrieves a causal distinction.

Another:

**What does a large positive standard reduction potential indicate?**

These propositions can support later chemical reasoning.

Retrieval-practice interventions have improved learning in general chemistry, showing that chemistry is not exempt from the broader testing effect.

But balancing a complicated redox reaction, predicting a mechanism, solving an equilibrium calculation, or interpreting spectroscopy is not maintained by remembering statements about those activities.

### Practical boundary

**Card stable chemical relationships, definitions, patterns, and key conditions. Practice calculations, mechanisms, structural interpretation, and experimental reasoning. Keep exhaustive property and safety tables external.**

---

## Biology

Biology contains a high density of declarative and relational knowledge, making it unusually compatible with retrieval cards.

Appropriate examples include:

**What is the role of helicase during DNA replication?**

**How does negative feedback contribute to homeostasis?**

**Why does competitive inhibition increase apparent \(K_m\) without changing \(V_{max}\)?**

Anatomy and physiology studies provide direct evidence for retrieval-based retention.

But biology also involves experimental design, interpreting figures, predicting the effect of perturbing a pathway, and reasoning across multiple levels of organization. Those tasks need application.

### Practical boundary

**Biology legitimately supports many cards, but “biology contains many facts” should not mutate into “all biology should become flashcards.”**

---

## Medicine

Medicine may be the clearest example of why both sides are necessary.

Cards are well suited to retaining anatomical relationships, disease mechanisms, pharmacological principles, characteristic findings, contraindications, and stable clinical criteria. Spaced retrieval has demonstrated benefits for clinical knowledge, including studies extending over many months.

Clinical reasoning is different.

A card might preserve that hyperkalemia can produce peaked T waves. Diagnosing and managing a patient with weakness, renal failure, several medications, and an abnormal ECG requires case reasoning.

Active retrieval from clinical cases can enhance diagnostic transfer to new cases.

Procedural skills are more literal still. Retaining CPR requires performing CPR-like tasks, not answering “How do you perform CPR?” on a card.

### Practical boundary

**Card the knowledge clinicians need available while reasoning. Use cases to retain diagnosis and management. Perform procedures to retain procedures. Look up volatile doses, guidelines, and detailed reference information when appropriate.**

---

## Programming and Computer Science

Programming mixes stable conceptual knowledge with highly perishable implementation detail.

Useful cards can preserve:

**Why can a hash table degrade from expected constant-time lookup?**

**What invariant does Dijkstra's algorithm rely on when finalizing a node?**

**What is the difference between deadlock and starvation?**

**When is memoization useful?**

But writing, debugging, tracing, profiling, designing, and refactoring code are performances.

Programming-education research on worked examples and Parsons problems demonstrates that activities differ substantially in what they ask learners to do. Rearranging existing program fragments, tracing programs, completing code, and generating code from scratch are related but non-identical demands.

A card stating the definition of a race condition is not practice diagnosing one.

Also, volatile syntax and library trivia are often ideal lookup material. Stable concepts usually have greater long-term value than remembering the seventh optional argument of an API that will be redesigned before the learner has finished reviewing the card.

### Practical boundary

**Card enduring concepts, invariants, complexity relationships, failure modes, and important semantics. Write and debug software to retain programming ability. Look up low-frequency syntax and rapidly changing API detail.**

---

## Economics

Economics contains definitions and compact model relations that are suitable for retrieval:

**What is opportunity cost?**

**Why does a binding price ceiling generate excess demand in the standard competitive model?**

**What distinguishes a movement along a demand curve from a shift in demand?**

But economics also requires model selection, comparative statics, causal inference, interpretation of data, and evaluation of policy under assumptions.

Problem-based economics research has explicitly examined content knowledge together with problem-solving outcomes, reflecting the domain's dual requirement for remembered economic concepts and applied reasoning.

### Practical boundary

**Card model components and conceptual relationships. Practice using models, interpreting evidence, and analyzing shocks or policies. Externalize current datasets, detailed schedules, and rapidly changing empirical values.**

---

## History

History contains many obvious candidates for retrieval:

chronological anchors, actors, institutions, treaties, political structures, concepts, and important causal propositions.

The month-long Butler and Roediger experiment using history lectures provides unusually direct evidence that recall testing can preserve this kind of historical content.

But historical expertise also requires constructing causal explanations, weighing sources, contextualizing claims, recognizing contingency, and comparing interpretations.

A card might ask:

**What fiscal problem contributed to the calling of the Estates-General in 1789?**

Useful.

A card asking:

**Was the French Revolution primarily caused by fiscal crisis, class conflict, political ideology, institutional dysfunction, or something else?**

cannot realistically encode the historical reasoning required by a serious answer.

### Practical boundary

**Card the factual and conceptual substrate of historical reasoning. Practice historical reasoning with sources, explanations, comparisons, and arguments.**

---

## Politics and Law

Political and legal knowledge often has a large retrievable substrate:

institutional powers, constitutional mechanisms, doctrinal tests, statutory concepts, electoral structures, holdings, and definitions.

These can be excellent cards when stable.

Legal performance, however, often requires issue spotting, analogizing cases, distinguishing precedent, interpreting language, applying standards to contested facts, and constructing arguments. Legal analogical reasoning depends on recognizing relevant similarities through domain knowledge rather than merely reciting the precedent.

Political reasoning similarly requires analyzing institutions in interaction, interpreting evidence, making causal claims, and evaluating policy tradeoffs.

### Practical boundary

**Card the machinery of the system. Practice using the machinery on cases, evidence, and arguments. Keep full statutes, changing regulations, current statistics, and exhaustive precedent details available for lookup.**

---

# 16. Flashcards Versus Exercises Is the Wrong Binary

One implication deserves emphasis.

Flashcards and exercises are not competitors for a single slot in a learning system.

They often maintain **different layers of competence**.

Suppose someone learns Bayesian inference.

Cards could preserve the definition of prior, likelihood, posterior, conditional independence, base-rate neglect, and perhaps several high-value relationships.

Problems preserve the ability to construct a Bayesian model from unfamiliar information.

Case analysis preserves the ability to decide whether Bayesian reasoning is useful in a messy real situation.

Reference tools preserve distribution tables, software documentation, and specialized numerical procedures.

The components support one another.

This also resolves an apparent contradiction in the retrieval-practice literature. Retrieval can improve conceptual learning and even transfer, while conventional cards can still be inadequate for complex problem solving.

Both are true.

The word **retrieval** covers very different tasks.

Retrieving a fact is retrieval practice.

Explaining a mechanism from memory is retrieval practice.

Solving a mathematics problem without looking at the solution is retrieval practice.

Diagnosing a patient case from remembered knowledge can involve retrieval practice.

Reconstructing a proof is retrieval practice.

Only the first two map naturally onto ordinary short Q&A flashcards.

The question is therefore not whether retrieval should occur. It almost certainly should.

The question is **what should be retrieved, in what form, to preserve the desired future capability?**

---

# 17. A Descriptive Decision Framework

For an already-understood piece of textbook knowledge, the first question should be:

### What would successful future use actually look like?

If successful future use is substantially **“I need to bring this fact, relationship, criterion, distinction, or principle to mind,”** Q&A retrieval is well matched.

If future use is **“I need to perform a sequence,”** preserve essential procedural knowledge with cards if useful, but practice the sequence.

If future use is **“I need to decide when this applies,”** cards can maintain conditions and cues, but mixed application exercises should preserve selection.

If future use is **“I need to construct an answer,”** such as a proof, program, diagnosis, causal explanation, or legal argument, then construction needs to be practiced.

If future use is **“I only need the precise value occasionally,”** ask whether there is any benefit to internalizing it at all.

The second question is:

### Does having this knowledge immediately available improve other thinking?

This is where foundational knowledge differs from trivia.

An identity repeatedly used inside larger mathematical reasoning may merit memorization even though it is derivable.

The boiling point of an obscure solvent at sixteen different pressures probably does not.

The third question is:

### Is the information stable?

Long-lived concepts are attractive retrieval targets.

Volatile regulatory, clinical, technological, or statistical detail requires caution. External reference may not merely be more efficient. It may be safer.

The fourth question is:

### Is the retrieval question exercising the same kind of cognition that matters later?

This is the strongest theme in the transfer literature.

A card about a rule does not necessarily train applying the rule.

A card about the steps of a procedure does not necessarily train performing the procedure.

A factual quiz does not automatically train higher-order application.

The closer the cognitive operation performed during retrieval is to the eventual use, the stronger the theoretical and empirical case for transfer.

---

# 18. The Most Important Boundary: Knowledge as an Operand Versus Knowledge as an Operation

A useful way to integrate the evidence is to distinguish between **operands of thought** and **operations of thought**.

An operand is something reasoning works *with*:

a definition, principle, relationship, criterion, theorem, mechanism, fact, invariant, model assumption, diagnostic feature, or historical event.

An operation is something reasoning *does*:

derive, calculate, compare, choose, debug, diagnose, design, prove, evaluate, interpret, construct, or argue.

This is not a recognized formal taxonomy in cognitive science. It is a synthesis of the reviewed evidence.

Q&A flashcards are particularly good at maintaining operands.

Exercises are particularly important for maintaining operations.

Complex professional knowledge requires both.

The distinction is not absolute because some operations can themselves be verbalized and retrieved. But verbal knowledge **about** an operation and competence **at** the operation should not be conflated.

Knowing the five steps of debugging is not debugging.

Knowing the definition of a proof by induction is not producing one.

Knowing the diagnostic criteria is not clinical diagnosis.

Knowing the legal test is not legal analysis.

Knowing a historical reasoning heuristic is not interpreting a primary source.

This is probably the most important reason not to indiscriminately turn textbooks into decks.

---

# 19. Evidence Strength and Important Uncertainties

The evidence is strongest for the proposition that effortful retrieval improves later access to learned information. This conclusion is supported by meta-analysis, laboratory research, classroom research, and multiple academic domains.

There is also substantial evidence that retrieval benefits can transfer beyond identical questions. Pan and Rickard's meta-analysis rules out the simplistic claim that retrieval practice merely teaches people to parrot exactly what was tested.

But evidence for **far transfer and complex professional performance** is substantially more conditional.

The strongest relevant finding may be that retrieval's cognitive level matters. Fact retrieval is highly effective for factual retention, but higher-order performance improves more reliably when practice itself requires higher-order retrieval or application.

There is also a meaningful unresolved literature surrounding complex problem solving. Some studies have found limited benefits from inserting retrieval practice into worked-example learning, while more recent studies explore conditions under which retrieval and examples can complement one another. It would be premature to reduce this literature to either “testing works for math” or “testing does not work for math.”

Another limitation is duration. Many retrieval studies measure retention over days or weeks. Month-long studies exist, and medicine contains several studies with substantially longer intervals, but evidence over years is far thinner than one would want for a system intended to maintain professional knowledge for decades.

The evidence also varies greatly by domain. Biology, medicine, science, and mathematics have relatively rich empirical literatures. Programming education has substantial research on problem types and worked examples but much less direct research on long-term Q&A flashcard use. Law, politics, and economics have still less research directly answering this specific question.

Finally, much published research concerns students rather than working professionals. The underlying memory mechanisms presumably do not vanish when someone receives a salary, but professional work changes the cost of lookup, frequency of use, stakes of errors, and forms of required transfer.

For those reasons, the framework in this paper should be treated as a synthesis of current evidence, not a proven optimal classification system.

---

# 20. Conclusion

The research does not support either extreme position about flashcards.

It is too restrictive to say:

**“Flashcards are only for arbitrary facts.”**

Retrieval practice can preserve conceptual relationships, causal explanations, applicability conditions, distinctions, and other meaningful knowledge. It can also produce some transfer beyond the exact retrieval task.

But it is equally mistaken to say:

**“If something matters, put it on a flashcard.”**

Successful recall of knowledge about a performance does not guarantee preservation of the performance itself. Mathematical problem solving, programming, diagnosis, procedural medicine, legal reasoning, historical analysis, model application, and scientific interpretation all contain operations that need to be exercised in their own right.

The best-supported boundary is therefore functional:

**Use Q&A retrieval cards to maintain knowledge whose valuable future form is ready mental availability.**

This includes not only isolated facts, but also important definitions, principles, relationships, causal mechanisms, conditions, discriminations, invariants, model assumptions, and frequently used cognitive building blocks.

**Use exercises and activities when the valuable future capability involves transforming or acting on knowledge.**

That includes solving, deriving, proving, programming, debugging, diagnosing, designing, comparing, interpreting evidence, evaluating cases, and constructing arguments.

**Use external reference when precise internal recall has little marginal value.**

Large tables, exhaustive catalogs, volatile specifications, rare constants, current datasets, detailed regulations, and similar material often belong there.

Across mathematics, physics, chemistry, biology, medicine, computer science, economics, history, politics, and law, the proportions differ. The principle does not change much.

The central question is not:

**“Is this a math fact or a biology fact?”**

It is:

**“What must this knowledge be able to do in the learner's mind six months or six years from now?”**

If the answer is **be recalled**, a Q&A flashcard is an unusually powerful tool.

If the answer is **be used to perform**, retrieval of the underlying knowledge is useful but insufficient.

If the answer is **be available if needed**, the smartest memory system may be the one that deliberately does not memorize it.

---

## Selected References

Agarwal, P. K. (2019). *Retrieval Practice & Bloom's Taxonomy: Do Students Need Fact Knowledge Before Higher Order Learning?* Journal of Educational Psychology, 111, 189–209.

Agarwal, P. K., Nunes, L. D., & Blunt, J. R. (2021). *Retrieval Practice Consistently Benefits Student Learning: A Systematic Review of Applied Research in Schools and Classrooms.* Educational Psychology Review.

Butler, A. C., & Roediger, H. L. (2007). *Testing improves long-term retention in a simulated classroom setting.* European Journal of Cognitive Psychology, 19, 514–527.

Dobson, J. L. (2013). *Retrieval practice is an efficient method of enhancing the retention of anatomy and physiology information.* Advances in Physiology Education.

Freeman, S., et al. (2014). *Active learning increases student performance in science, engineering, and mathematics.* Proceedings of the National Academy of Sciences.

Gjerde, V. (2022). *Problem solving in basic physics: Effective self-explanations based on four elements with support from retrieval practice.* Physical Review Physics Education Research.

Karpicke, J. D., & Blunt, J. R. (2011). *Retrieval practice produces more learning than elaborative studying with concept mapping.* Science, 331, 772–775.

Kerfoot, B. P., et al. (2007). *Spaced education improves the retention of clinical knowledge by medical students: a randomised controlled trial.* Medical Education, 41, 23–31.

Kromann, C. B., et al. (2010). *The testing effect on skills learning might last 6 months.* Advances in Health Sciences Education.

Pan, S. C., & Rickard, T. C. (2018). *Transfer of test-enhanced learning: Meta-analytic review and synthesis.* Psychological Bulletin, 144, 710–756.

Risko, E. F., & Gilbert, S. J. (2016). *Cognitive Offloading.* Trends in Cognitive Sciences, 20, 676–688.

Rohrer, D., et al. (2014). *The benefit of interleaved mathematics practice is not limited to superficially similar kinds of problems.* Psychonomic Bulletin & Review.

Rowland, C. A. (2014). *The effect of testing versus restudy on retention: A meta-analytic review of the testing effect.* Psychological Bulletin, 140, 1432–1463.

Sheldon, S., et al. (2023). *Learning strategy impacts medical diagnostic reasoning in early learners.*