(* Matching verification (in progress)

   See https://github.com/Comcast/sheens#pattern-matching.

   The start of a verification of that pattern matching algorithm.
   This file is written for the Coq Proof Assistant:
   https://coq.inria.fr/.

   The matching here gives a set (list) of returned bindings, but
   currently the matching doesn't support arrays (as sets).

   Status: A work in progress.

 *)

Require Import String Bool Arith List.

(* for jsCoq: Comments "pkgs: coq-arith".  Then From Coq Require Import String. *)

(* Require Import Cpdt.CpdtTactics. 
   Set Implicit Arguments.
   Set Asymmetric Patterns. *)

Module patmatch.

  (* We define our own association list to help with
     well-foundedness arguments for important functions. *)
  
  Section alists.

    (* Our association lists. *)
    Definition new_alist {T:Set} : (list (string*T)) := nil.

    (* Add a pair to an alist. *)
    Fixpoint acons {T:Set} (a: list (string*T)) (k:string) (v:T) : list (string*T) :=
      match a with
      | nil => (k,v)::nil
      | (k',v')::more =>
        if string_dec k k'
        then acons more k v
        else (k',v')::(acons more k v)
      end.

    (* Get the value for a given key. *)
    Fixpoint assoc {T:Set} (a: list (string*T)) (k: string) : option T :=
      match a with
      | nil => None
      | (p,v)::more =>
        if string_dec k p then Some v else assoc more k
      end.

  End alists.

  Section patterns_and_messages.

    (* Patterns *)
    Inductive pat : Set :=
    | PStr : string -> pat
    | Var : string -> pat
    | PMap : list (string*pat) -> pat.

    (* Messages *)
    Inductive msg : Set :=
    | Str : string -> msg
    | Map : list (string*msg) -> msg.

    Definition bindings := list (string*msg).

  End patterns_and_messages.

  Section matching.

    (* Is the first message is a sub-message of the second message? *)
    Fixpoint submsg (p:msg) (m:msg) : bool :=
      match p, m with
      | Str s, Str s' =>
        if string_dec s s' then true else false
      | Map kvs, Map mm =>
        let fix f kvs :=
            match kvs with
            | nil => true
            | (k,v)::kvs' =>
              match assoc mm k with
              | None => false
              | Some v' =>
                if submsg v v'
                then f kvs'
                else false
              end
            end
        in f kvs
      |  _, _ => false
      end.

    Definition apply_app {A:Type} (lsts:list (list A)) :=
      fold_right (fun x y => x ++ y) nil lsts.

    (* The main function. *)
    Fixpoint patmatch (p:pat) (bs:bindings) (m:msg) : list bindings :=
      match p, m with
      | PStr ps, Str ms =>
        if string_dec ps ms then bs::nil else nil
      | Var v, _ =>
        match assoc bs v with
        | None => (acons bs v m)::nil
        | Some x => if submsg x m then bs::nil else nil
        end
      | PMap pm, Map mm =>
        let fix f pm bs :=
            match pm with
            | nil => bs::nil
            | (k,v)::pm' =>
              match assoc mm k with
              | None => nil
              | Some v' =>
                apply_app
                  (map (fun (bs:bindings) =>
                          (f pm' bs))
                       (patmatch v bs v'))
              end
            end
        in f pm bs
      | _, _ => nil
      end.

  End matching.

  (* Just some computations to take a look around. *)
  Section patmatch_tests.

    Compute patmatch (PStr "chips") nil (Str "chips") .
    
    Compute let p := (PMap nil) in
            let m := (Map nil) in
            patmatch p nil m.

    Compute let p := (PMap (acons nil "likes" (PStr "tacos"))) in
            let m := (Map (acons nil "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons nil "likes" (Var "x"))) in
            let m := (Map (acons nil "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons nil "likes" (Var "x"))) in
            let m := (Map (acons (acons nil "wants" (Str "chips"))
                                 "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons (acons nil "wants" (Var "y"))
                                  "likes" (Var "x"))) in
            let m := (Map (acons (acons nil "wants" (Str "chips"))
                                 "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons (acons (acons nil "needs" (Var "y"))
                                         "wants" (Var "y"))
                                  "likes" (Var "x"))) in
            let m := (Map (acons (acons (acons nil "needs" (Str "chips"))
                                        "wants" (Str "chips"))
                                 "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons (acons (acons nil "needs" (Var "y"))
                                         "wants" (Var "y"))
                                  "likes" (Var "x"))) in
            let m := (Map (acons (acons (acons nil "needs" (Str "queso"))
                                        "wants" (Str "chips"))
                                 "likes" (Str "tacos"))) in
            patmatch p nil m.


  End patmatch_tests.

  Section verification.

    (* assoc after an acons does what you'd expect. *)
    Remark acons_assoc :
      forall m:list (string*msg),
      forall k:string,
      forall v:msg,
        (assoc (acons m k v) k) = Some v.
      intros.
      induction m.
      simpl.
      destruct (string_dec k k).
      reflexivity.
      intuition.
      simpl.
      induction a.
      destruct (string_dec k a).
      assumption.
      simpl.
      destruct (string_dec k a).
      contradiction.
      assumption.
    Defined.

    (* Do the bindings applied to the pattern give a submsg of the
       message? *)
    Fixpoint psubmsg (p:pat) (bs:bindings) (m:msg) : bool :=
      match p, m with
      | PStr ps, Str ms =>
        if string_dec ps ms then true else false
      | Var k, _ =>
        match assoc bs k with
        | None => false
        | Some v => submsg v m
        end
      | PMap pm, Map mm =>
        let fix f pm :=
            match pm with
            | nil => true
            | (k,v)::pm' =>
              match assoc mm k with
              | None => false
              | Some v' =>
                if psubmsg v bs v'
                then f pm'
                else false
              end
            end
        in f pm
      |  _, _ => false
      end.

    (* Submsg is reflexive. *)
    Lemma submsg_refl : forall x:msg, submsg x x = true.
    Admitted.

    (* Using bindings from a map match results in a submsg. *)
    Lemma patmatch_submsg_maps :
      forall l: list (string * pat),
      forall l0 : list (string * msg),
      forall bs : bindings,
        In bs (patmatch (PMap l) nil (Map l0)) ->
        psubmsg (PMap l) bs (Map l0) = true.
    Admitted.

    (* Big theorem #1: Bindings from a patmatch given a submsg. *)
    Theorem patmatch_submsg :
      forall p:pat,
      forall m:msg,
      forall bs:bindings,
        In bs (patmatch p nil m) ->
        psubmsg p bs m = true.
    Proof.
      intros.
      induction p.
      {
        induction m.
        {
          unfold psubmsg.
          destruct (string_dec s s0).
          {
            reflexivity.
          }
          {
            unfold patmatch in H.
            destruct (string_dec s s0).
            {
              unfold In in H.
              intuition.
            }
            {
              unfold In in H.
              contradiction.
            }
          }
        }
        {
          unfold psubmsg.
          unfold patmatch in H.
          unfold In in H.
          contradiction.
        }
      }
      {
        induction m.
        {
          unfold psubmsg.
          unfold patmatch in H.
          unfold assoc in H.
          unfold acons in H.
          unfold In in H.
          intuition.
          rewrite <- H0.
          simpl.
          destruct (string_dec s s).
          {
            simpl.
            destruct (string_dec s0 s0).
            reflexivity.
            intuition.
          }
          {
            intuition.
          }
        }
        {
          unfold psubmsg.
          unfold patmatch in H.
          unfold assoc in H.
          unfold acons in H.
          unfold In in H.
          intuition.
          rewrite <- H0.
          unfold assoc.
          destruct (string_dec s s).
          {
            apply submsg_refl.
          }
          {
            intuition.
          }
        }
      }
      {
        induction m.
        {
          unfold psubmsg.
          unfold patmatch in H.
          intuition.
        }
        {
          apply patmatch_submsg_maps in H.
          trivial.
        }
      }
    Qed.

    (* Big theorem #2: If some bindings give a submsg, then patmatch
    should find those bindings. *)
    Theorem submsg_patmatch :
      forall p:pat,
      forall m:msg,
      forall bs:bindings,
        psubmsg p bs m = true ->
        In bs (patmatch p nil m).
      Admitted.
    
  End verification.

End patmatch.
